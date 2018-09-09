package dsm

// routines to interact with etcdv3:
import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3lock"
	"go.etcd.io/etcd/etcdserver/api/v3lock/v3lockpb"
	"log"
	"time"
)

const (
	DIAL_TIMEOUT = 5 * time.Second
	LOCK_DURATION = 10 * time.Second
	LOCK_REQUEST_TIMEOUT = 100 * time.Millisecond // max time to wait on a lock acquisition
	REQUEST_TIMEOUT = 10 * time.Second
)

// client wrapper on etcdv3 client
type Client struct {
	client *clientv3.Client
}

// create a new connection to the etcd cluster
func NewClient(url *string) *Client{

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{*url},
		DialTimeout: DIAL_TIMEOUT,
	})
	if err != nil {
		log.Fatal("Failed to connect to etcd cluster at: ", url)
	}

	log.Println("Connected to etcd cluster at ", *url)
	return &Client{client: client}
}

// shutdown the client connection
func (c *Client) Close() {
	log.Println("Disconnected from etcd cluster")
	c.client.Close()
}

// delete the given key
func (c *Client) Delete(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()

	_, err := c.client.Delete(ctx, key)
	if err != nil {
		log.Print("Failed to delete key ", key, ", err: ", err)
	}
}


// Get the value of a key
func (c *Client) Get(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()

	getResponse, err := c.client.Get(ctx, key)

	var value []byte
	value = nil

	if err == nil {
		value = getResponse.Kvs[0].Value
	}

	return value, err
}

// Get values matching a given prefix - return a map of key to value
func (c *Client) GetByPrefix(prefix string) ([][]byte, error) {

	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()

	var values [][]byte
	values = nil

	getResponse, err := c.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err == nil {
		numValues := len(getResponse.Kvs)
		values = make([][]byte, numValues)

		for i, kv := range getResponse.Kvs {
			values[i] = kv.Value
		}
	}

	return values, err
}

// Create/acquire a shared lock with the given name.  Return true if lock was successfully acquired.
func (c *Client) Lock(name string) (success bool, unlockFunction func()) {

	lockName := "lock-" + name

	lockServer := v3lock.NewLockServer(c.client)

	// lock duration puts a bound on how long a non-responsive client can hold a lock
	leaseID := c.newLease(LOCK_DURATION)
	lockRequest := v3lockpb.LockRequest{Name: []byte(lockName), Lease: int64(leaseID)}

	ctx, cancel := context.WithTimeout(context.Background(), LOCK_REQUEST_TIMEOUT)
	defer cancel()

	log.Print("Locking ", lockName)

	lockResponse, err := lockServer.Lock(ctx, &lockRequest)

	if err != nil {
		log.Print("Error locking ", lockName, ": ", err)
		return false, func() {}
	} else {
		unlockKey := lockResponse.Key

		unlockFunction := func() {
			c.Unlock(unlockKey)

			// cancel associated lease
			c.revokeLease(leaseID)
		}
		return true, unlockFunction
	}
}


// Set a key, with a limited expiration
func (c *Client) Put(key string, value string, leaseSeconds time.Duration) error {

	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()

	getResponse, err := c.client.Get(ctx, key)
	if err != nil {
		log.Fatal("Failed to get key ", key)
	}

	keyExists := len(getResponse.Kvs) == 1
	if keyExists {
		oldLease := clientv3.LeaseID(getResponse.Kvs[0].Lease)
		defer c.revokeLease(oldLease)
	}

	// create a new lease
	leaseID := c.newLease(leaseSeconds)

	_, err = c.client.Put(ctx, key, value, clientv3.WithLease(leaseID))
	return err
}

// unlock the given lock
func (c *Client) Unlock(unlockKey []byte) {
	lockServer := v3lock.NewLockServer(c.client)

	ctx, cancel := context.WithTimeout(context.Background(), LOCK_REQUEST_TIMEOUT)
	defer cancel()


	unlockRequest := v3lockpb.UnlockRequest{Key: unlockKey}

	unlockKeyString := string(unlockKey)
	log.Print("Unlocking ", unlockKeyString)

	_, err := lockServer.Unlock(ctx, &unlockRequest)
	if err != nil {
		// failure to unlock might mean the lease ran out, so
		log.Fatal("Failed to unlock ", unlockKeyString, ": ", err)
	}
}

// lease creation wrapper
func (c *Client) newLease(secs time.Duration) clientv3.LeaseID {

	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()

	lease := clientv3.NewLease(c.client)
	leaseGrantResponse, err := lease.Grant(ctx, 2000)

	if err != nil {
		log.Fatal("Failed to grant lease: ", err)
	}

	return leaseGrantResponse.ID
}

// lease revocation wrapper
func (c *Client) revokeLease(leaseId clientv3.LeaseID) {

	ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT)
	defer cancel()

	lease := clientv3.NewLease(c.client)
	_, err := lease.Revoke(ctx, leaseId)

	if err != nil {
		log.Print("Failed to revoke lease: ", err)
	}
}


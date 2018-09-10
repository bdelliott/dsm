package dsm

import "time"

// contract for backend distributed state machinery to implement
type Client interface {
	Close()
	Delete(key string)
	Get(key string) ([]byte, error)
	GetByPrefix(prefix string) ([][]byte, error)
	Lock(name string) (success bool, unlockFunction func())
	Put(key string, value string, leaseSeconds time.Duration) error
	Unlock(unlockKey []byte)
}

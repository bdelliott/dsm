package dsm

import (
	"strings"
	"time"
)

// test implementation of Client interface

type TestClient struct {
	machines map[string][]byte
}

func (c TestClient) Close() {}

func (c TestClient) Delete (key string) {}

func (c TestClient) Get(key string) ([]byte, error) {
	return c.machines[key], nil
}

func (c TestClient) GetByPrefix(prefix string) ([][]byte, error) {

	var results [][]byte

	for k, v := range c.machines {
		if strings.HasPrefix(k, MACHINE_PREFIX) {
			results = append(results, v)
		}
	}
	return results, nil
}

func (c TestClient) Lock(name string) (success bool, unlockFunction func()) { return true, func() {}}

func (c TestClient) Put(key string, value string, leaseSeconds time.Duration) error {
	c.machines[key] = []byte(value)
	return nil;
}

func (c TestClient) Unlock(unlockKey []byte) {}

func NewTestClient() TestClient {
	c := TestClient{}
	c.machines = make(map[string][]byte)
	return c
}
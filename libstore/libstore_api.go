package libstore

import (
	"achadha235/p3/rpc/storagerpc"
	"hash/fnv"
)

// Libstore defines the set of methods that a StockServer can call on its
// local cache.
type Libstore interface {
	Get(key string) (string, error)
	Put(key, value string) error
	GetList(key string) ([]string, error)
	AppendToList(key, newItem string) error
	RemoveFromList(key, removeItem string) error
	Transact(name storagerpc.TransactionType, data string) (storagerpc.TransactionStatus, error)
}

// StoreHash hashes a string key and returns a 32-bit integer. This function
// is provided here so that all implementations use the same hashing mechanism
// (both the Libstore and StorageServer should use this function to hash keys).
func StoreHash(key string) uint32 {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return hasher.Sum32()
}

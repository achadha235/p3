package libstore

import (
	"achadha235/p3/rpc/storagerpc"
)

// Libstore defines the set of methods that a StockServer can call on its
// local cache.
type Libstore interface {
	Get(key string) (string, error)
	Put(key, value string) error
	GetList(key string) ([]string, error)
	AppendToList(key, newItem string) error
	RemoveFromList(key, removeItem string) error
}

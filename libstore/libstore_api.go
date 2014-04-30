package libstore

import (
	"achadha235/p3/datatypes"
	"achadha235/p3/rpc/storagerpc"
)

// Libstore defines the set of methods that a StockServer can call on its
// local cache.
type Libstore interface {
	Get(key string) (string, storagerpc.Status, error)
	Transact(name datatypes.TransactionType, data *datatypes.DataArgs) (datatypes.Status, error)
}

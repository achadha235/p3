package coordinator

import (
	"github.com/achadha235/p3/datatypes"
)

type Coordinator interface {
	PerformTransaction(name datatypes.TransactionType, data datatypes.DataArgs) (datatypes.Status, error)
}

package cohort

import (
	"github.com/achadha235/p3/rpc/storagerpc"
)

type Cohort interface {
	ExecuteTransaction(*storagerpc.TransactionArgs, *storagerpc.TransactionReply) error // This should be defined by the application
	TransactionIsLegal(*storagerpc.TransactionArgs, *storagerpc.TransactionReply) bool  //
	RollBack(*storagerpc.TransactionArgs, *storagerpc.TransactionReply) error

	RegisterServer(*storagerpc.RegisterServerArgs, *storagerpc.RegisterServerReply) error
	GetServers(*storagerpc.GetServersArgs, *storagerpc.GetServersReply) error

	Prepare(*storagerpc.PrepareArgs, *storagerpc.PrepareReply) error
	Commit(*storagerpc.CommitArgs, *storagerpc.CommitReply) error
	Get(*storagerpc.GetArgs, *storagerpc.GetReply) error
}

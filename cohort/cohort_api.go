package cohort

import (
	"github.com/achadha235/p3/rpc/storagerpc"
)

type CohortStorageServer interface {
	RegisterServer(*storagerpc.RegisterArgs, *storagerpc.RegisterReply) error
	GetServers(*storagerpc.GetServersArgs, *storagerpc.GetServersReply) error

	Prepare(*storagerpc.PrepareArgs, *storagerpc.PrepareReply) error
	Commit(*storagerpc.CommitArgs, *storagerpc.CommitReply) error
	Get(*storagerpc.GetArgs, *storagerpc.GetReply) error
}

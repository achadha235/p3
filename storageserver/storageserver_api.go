package storageserver

// StorageServer defines the RPC interface for the 2PC-based storage consumed by StockServer
import "github.com/achadha235/p3/rpc/storagerpc"

type StorageServer interface {
	
	// RegisterServer adds a storage server to the cohort. Common to all 
	GetServers(*storagerpc.GetServers) //
	VoteMaster(*storagerpc.VoteArgs, *storageserver.VoteReply)
	Get(*storagerpc.GetArgs, *storagerpc.GetReply) error
	Put(*storagerpc.PutArgs, *storagerpc.PutReply) error


	// Only if currently master register a new slave to the 2PC cohort 
	RegisterServer(*storagerpc.RegisterArgs, *storagerpc.RegisterReply) error
	// Prepare all servers to commit
	PrepareServers(*storagerpc.PrepareArgs, *storage.PrepareReply) error
	// Request all servers to commit
	CommitServers(*storagerpc.CommitArgs, *storage.CommitReply) error
}


package storagerpc

type StorageServer interface {
	RegisterServer(*RegisterServerArgs, *RegisterServerReply) error
	GetServers(*GetServersArgs, *GetServersReply) error
	Get(*GetArgs, *GetReply) error
	Put(*PutArgs, *PutReply) error
}

type RemoteMasterStorageServer interface {
	RegisterServer(*RegisterServerArgs, *RegisterServerReply) error
	Propose(*ProposeArgs, *ProposeReply) error
}

type RemoteCohortStorageServer interface { // This should be RemoteCohortServer
	Prepare(*PrepareArgs, *PrepareReply) error
	ExecuteTransaction(*TransactionArgs, *TransactionReply) error // This should be defined by the application
	Commit(*CommitArgs, *CommitReply) error
	Get(*GetArgs, *GetReply) error
}

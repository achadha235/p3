package storagerpc

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
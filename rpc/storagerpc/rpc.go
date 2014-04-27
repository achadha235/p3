package storagerpc

type MasterStorageServer interface {
	RegisterServer(*RegisterServerArgs, *RegisterServerReply) error 
	Propose(*ProposeArgs, *ProposeReply) error 
}

type CohortStorageServer interface {
	Prepare(*PrepareArgs, *PrepareReply) error 
	Commit(*CommitArgs, *CommitReply) error 
	Get(*GetArgs, *GetReply) error 
	Execute(*ExecuteArgs, *ExecuteReply) error 
}
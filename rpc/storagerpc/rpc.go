package storagerpc

type RemoteMasterServer interface {
	RegisterServer(*RegisterServerArgs, *RegisterServerReply) error
	Propose(*ProposeArgs, *ProposeReply) error
}

type RemoteCohortServer interface { // This should be RemoteCohortServer
	Prepare(*PrepareArgs, *PrepareReply) error
	ExecuteTransaction(*TransactionArgs, *TransactionReply) error // This should be defined by the application
	Commit(*CommitArgs, *CommitReply) error
	Get(*GetArgs, *GetReply) error
	GetServers(*GetServersArgs, *GetServersReply) error
}

type CohortServer struct {
	RemoteCohortServer
}

func Wrap(s RemoteCohortServer) RemoteCohortServer {
	return &CohortServer{s}
}

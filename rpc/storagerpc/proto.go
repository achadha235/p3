package storagerpc

type RemoteStorageServer interface {
	RegisterServer(*RegisterArgs, *RegisterReply) error 
	GetServers(*GetServersArgs, *GetServersReply) error
	Get(*GetArgs, *GetReply) error
	Put(*PutArgs, *PutReply) error
}

type StorageServer struct {
	RemoteStorageServer
}
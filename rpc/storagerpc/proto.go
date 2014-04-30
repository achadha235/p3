package storagerpc

import (
	"github.com/achadha235/p3/datatypes"
)

type Status int

const (
	OK Status = iota
	NotReady
	KeyNotFound
)

type CommitStatus int

const (
	Commit = iota
	Rollback
)

type Node struct {
	NodeId   uint32
	HostPort string
	Master   bool
}

type TransactionType int

const (
	CreateUser = iota
	CreateTeam
	JoinTeam
	LeaveTeam
	MakeTransaction
)

type TransactionArgs struct {
	TransactionId int
	Method        TransactionType
	Data          TransactionData
}

type TransactionReply struct {
	Status datatypes.Status
	Error  error
}

type TransactionData struct {
	jsonString string
}

type RegisterArgs struct {
	ServerInfo Node
}

type RegisterReply struct {
	Status  Status
	Servers []Node
}

type GetServersArgs struct {
}

type GetServersReply struct {
	Status  Status
	Servers []Node
}

type PrepareArgs struct {
	TransactionId int
	Name          datatypes.OperationType
	Data          datatypes.DataArgs
}

type PrepareReply struct {
	Status datatypes.Status
}

type CommitArgs struct {
	TransactionId int
	Status        CommitStatus
}

type CommitReply struct {
}

type GetArgs struct {
	Key string
}

type GetReply struct {
	Key           string
	Value         string
	Status        datatypes.Status
	StorageStatus Status
}

type PutArgs struct {
	Key   string
	Value string
}

type PutReply struct {
}

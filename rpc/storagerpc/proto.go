package storagerpc

import (
	"github.com/achadha235/p3/datatypes"
	"net/rpc"
)

type Status int

const (
	OK Status = iota + 1
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
	Client   *rpc.Client
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

type RegisterServerArgs struct {
}

type RegisterServerReply struct {
	Status int
	NodeId int
}

type GetServersArgs struct {
}

type GetServersReply struct {
	Status  datatypes.Status
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

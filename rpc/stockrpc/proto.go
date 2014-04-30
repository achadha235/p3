// This file contains constants and arguments used to perform RPC calls between
// a StockClient and StockServer

package stockrpc

import (
	"github.com/achadha235/p3/datatypes"
)

// v0: Portfolio actions supported : Buy/Sell

type LoginUserArgs struct {
	UserID, Password string
}

type LoginUserReply struct {
	Status     datatypes.Status
	SessionKey []byte
}

type CreateUserArgs struct {
	UserID, Password string
}

type CreateUserReply struct {
	Status datatypes.Status
}

type CreateTeamArgs struct {
	TeamID, Password string
	SessionKey       []byte
}

type CreateTeamReply struct {
	Status datatypes.Status
}

type JoinTeamArgs struct {
	TeamID, Password string
	SessionKey       []byte
}

type JoinTeamReply struct {
	Status datatypes.Status
}

type LeaveTeamArgs struct {
	TeamID     string
	SessionKey []byte
}

type LeaveTeamReply struct {
	Status datatypes.Status
}

// Request a transaction to be made
type MakeTransactionArgs struct {
	Requests   []datatypes.Request
	SessionKey []byte
}

type MakeTransactionReply struct {
	Status datatypes.Status
}

// Get the portfolio for a team
type GetPortfolioArgs struct {
	TeamID string
}

type GetPortfolioReply struct {
	Stocks []datatypes.Holding
	Status datatypes.Status
}

type GetPriceArgs struct {
	Ticker string
}

type GetPriceReply struct {
	Price  uint64 // Price is in cents
	Status datatypes.Status
}

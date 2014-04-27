// This file contains constants and arguments used to perform RPC calls between
// a StockClient and StockServer

package stockrpc

import (
	"time"
)

// Status represents status of an RPC's reply
type Status int

// v0: Portfolio actions supported : Buy/Sell

const (
	OK                   Status = iota + 1 // RPC was successful
	NoSuchUser                             // Specified user does not exist
	NoSuchTeam                             // Specified team does not exist
	NoSuchTicker                           // Specified stock does not exist
	NoSuchAction                           // Specified action not supported
	NoSuchSession                          // Specified session key does not exist
	InsufficientQuantity                   // Desired action cannot be fulfilled; lack of money/shares
	Exists                                 // User/team already exists or user is already on team
	PermissionDenied                       // User does not have permission to do the task
)

// struct used to represent the possession of shares of a stock for teams
type Holding struct {
	ticker   string
	quantity uint64
	acquired time.Time
}

// request structure for a single action (buy/sell in v0)
type Request struct {
	action, teamID, ticker string
	quantity               int
}

// struct used to represent a user
type User struct {
	userID string
	hashPW string   // hashed PW
	teams  []string // list of team IDs that the user is on
}

// struct used to represent a team
type Team struct {
	users    []string  // list of userIDs of users that are on the team
	hashPW   string    // hashed PW
	balance  uint64    // balance in cents
	holdings []Holding // list of holding IDs
}

// struct for args to JoinTeam Transaction
type UserTeamData struct {
	userID string
	teamID string
}

type Ticker struct {
	price uint64
}

type LoginUserArgs struct {
	UserID, Password string
}

type LoginUserReply struct {
	Status     Status
	SessionKey []byte
}

type CreateUserArgs struct {
	UserID, Password string
}

type CreateUserReply struct {
	Status Status
}

type CreateTeamArgs struct {
	TeamID, Password string
	SessionKey       []byte
}

type CreateTeamReply struct {
	Status Status
}

type JoinTeamArgs struct {
	TeamID, Password string
	SessionKey       []byte
}

type JoinTeamReply struct {
	Status Status
}

type LeaveTeamArgs struct {
	TeamID     string
	SessionKey []byte
}

type LeaveTeamReply struct {
	Status Status
}

// Request a transaction to be made
type MakeTransactionArgs struct {
	Requests   []Request
	SessionKey []byte
}

type MakeTransactionReply struct {
	Status Status
}

// Get the portfolio for a team
type GetPortfolioArgs struct {
	TeamID string
}

type GetPortfolioReply struct {
	Stocks []Holding
	Status Status
}

type GetPriceArgs struct {
	Ticker string
}

type GetPriceReply struct {
	Price  int64 // Price is in cents
	Status Status
}

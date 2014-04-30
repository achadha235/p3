/* This package contains all info for statuses and
   datatypes */
package datatypes

import (
	"time"
)

// Status represents status of an RPC's reply
type Status int
type TransactionType int
type OperationType int

const (
	OK                   Status = iota + 1 // Action was successful
	NoSuchUser                             // Specified user does not exist
	NoSuchTeam                             // Specified team does not exist
	NoSuchTicker                           // Specified stock does not exist
	NoSuchAction                           // Specified action not supported
	NoSuchSession                          // Specified session key does not exist
	NoSuchHolding                          // Specified holding not found
	InsufficientQuantity                   // Desired action cannot be fulfilled; lack of money/shares
	Exists                                 // User/team already exists or user is already on team
	PermissionDenied                       // User does not have permission to do the task
	BadData                                // Data is not properly stored or system is corrupted
)

const (
	CreateUser TransactionType = iota + 1
	CreateTeam
	JoinTeam
	LeaveTeam
	MakeTransaction
)

const (
	AddUser OperationType = iota + 1
	AddTeam
	AddUserToTeamList
	AddTeamToUserList
	RemoveUserFromTeamList
	RemoveTeamFromUserList
	Buy
	Sell
)

// struct used to represent the possession of shares of a stock for teams
type Holding struct {
	Ticker   string
	Quantity uint64
	Acquired time.Time
}

// request structure for a single action (buy/sell in v0)
type Request struct {
	Action, TeamID, Ticker string
	Quantity               int
}

// struct used to represent a user
type User struct {
	UserID string
	HashPW string   // hashed PW
	Teams  []string // list of team IDs that the user is on
}

// struct used to represent a team
type Team struct {
	TeamID   string
	Users    []string // list of userIDs of users that are on the team
	HashPW   string   // hashed PW
	Balance  uint64   // balance in cents
	Holdings []string // list of holding IDs
}

type Ticker struct {
	Price uint64
}

// Arguments for the RPC call for operations on CohortServers
type DataArgs struct {
	User     User
	Team     Team
	Pw       string
	Requests []Request
}

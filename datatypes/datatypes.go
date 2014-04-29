/* This package contains all info for statuses and
   datatypes */
package datatypes

import (
	"time"
)

// Status represents status of an RPC's reply
type Status int
type TransactionType int

const (
	CreateUser TransactionType = iota + 1
	CreateTeam
	JoinTeam
	LeaveTeam
	MakeTransaction
)

const (
	OK                   Status = iota + 1 // Action was successful
	NoSuchUser                             // Specified user does not exist
	NoSuchTeam                             // Specified team does not exist
	NoSuchTicker                           // Specified stock does not exist
	NoSuchAction                           // Specified action not supported
	NoSuchSession                          // Specified session key does not exist
	InsufficientQuantity                   // Desired action cannot be fulfilled; lack of money/shares
	Exists                                 // User/team already exists or user is already on team
	PermissionDenied                       // User does not have permission to do the task
	BadData                                // Data is not properly stored or is corrupted
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
	teamID   string
	users    []string  // list of userIDs of users that are on the team
	hashPW   string    // hashed PW
	balance  uint64    // balance in cents
	holdings []Holding // list of holding IDs
}

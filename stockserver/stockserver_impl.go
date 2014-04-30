package stockserver

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/json"
	"errors"
	"github.com/achadha235/p3/datatypes"
	"github.com/achadha235/p3/libstore"
	"github.com/achadha235/p3/rpc/stockrpc"
	"github.com/achadha235/p3/rpc/storagerpc"
	"github.com/achadha235/p3/util"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const (
	DefaultStartAmount = 10000000 // starting amount (in cents)
)

type stockServer struct {
	server     *rpc.Server
	ls         libstore.Libstore
	sessionMap map[string]string // map from sessionKey to userID for that session
}

// NewStockServer creates, starts, and returns a new StockServer.
// masterHostPort is the coordinator's host:port and myHostPort is the
// port that the StockServer should listen on for StockClient calls

func NewStockServer(masterHostPort, myHostPort string) (StockServer, error) {
	s := rpc.NewServer()

	ls, err := libstore.NewLibstore(masterHostPort, myHostPort)
	if err != nil {
		return nil, err
	}

	sMap := make(map[string]string)
	ss := &stockServer{
		server:     s,
		ls:         ls,
		sessionMap: sMap,
	}

	listener, err := net.Listen("tcp", myHostPort)
	if err != nil {
		return nil, err
	}

	err = rpc.RegisterName("StockServer", stockrpc.Wrap(ss))
	if err != nil {
		log.Println("Failed to register StockServer RPC")
		return nil, err
	}

	rpc.HandleHTTP()
	go http.Serve(listener, nil)

	return ss, nil
}

// given a sessionKey, return the associated userID.
// Returns non-nil error if sessionKey does not exist
func (ss *stockServer) RetrieveSession(sessionKey []byte) (string, error) {
	if userID, ok := ss.sessionMap[string(sessionKey)]; !ok {
		return "", errors.New("No such sessionKey")
	} else {
		return userID, nil
	}
}

// LoginUser checks that the user exists and returns a unique session key for that user
func (ss *stockServer) LoginUser(args *stockrpc.LoginUserArgs, reply *stockrpc.LoginUserReply) error {
	key := util.CreateUserKey(args.UserID)
	// check if the user exists
	encodedUser, status, err := ss.ls.Get(key)
	if err != nil {
		log.Println("Get err on user login: ", err)
		reply.Status = datatypes.BadData
		return nil
	} else if status != storagerpc.OK {
		reply.Status = datatypes.NoSuchUser
		return nil
	}

	var user datatypes.User
	err = json.Unmarshal([]byte(encodedUser), &user)
	if err != nil {
		log.Println("fail to unmarshal: ", err)
		reply.Status = datatypes.BadData
		return nil
	}

	// check if user pw is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.HashPW), []byte(args.Password))
	if err != nil {
		reply.Status = datatypes.PermissionDenied
		return nil
	}

	// create session key
	sessionID := args.UserID + time.Now().String()
	sessionKey, err := bcrypt.GenerateFromPassword([]byte(sessionID), bcrypt.DefaultCost)
	if err != nil {
		log.Println("bad session data", err)

		reply.Status = datatypes.BadData
		return nil
	}

	// save the session key for current user
	ss.sessionMap[string(sessionKey)] = args.UserID

	reply.SessionKey = sessionKey
	reply.Status = datatypes.OK
	return nil
}

// CreateUser adds a user to the game, or returns Exists if userID is already in use
func (ss *stockServer) CreateUser(args *stockrpc.CreateUserArgs, reply *stockrpc.CreateUserReply) error {

	hashed, err := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)
	if err != nil {
		reply.Status = datatypes.BadData
		return nil
	}

	// create user object
	user := datatypes.User{
		UserID: args.UserID,
		HashPW: string(hashed),
		Teams:  make([]string, 0),
	}

	data := &datatypes.DataArgs{User: user}

	status, err := ss.ls.Transact(datatypes.CreateUser, data)
	if err != nil {
		reply.Status = status
		return err
	}

	reply.Status = status
	return nil
}

// CreateTeam adds a team to the game, or returns Exists if teamID is already in use
func (ss *stockServer) CreateTeam(args *stockrpc.CreateTeamArgs, reply *stockrpc.CreateTeamReply) error {
	// lookup userID based on session
	userID, err := ss.RetrieveSession(args.SessionKey)
	if err != nil {
		reply.Status = datatypes.NoSuchSession
		return nil
	}

	// Add user to the team he created
	userList := make([]string, 0)
	userList = append(userList, userID)

	// Create team pw
	hashed, err := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)
	if err != nil {
		reply.Status = datatypes.BadData
		return nil
	}

	team := datatypes.Team{
		TeamID:   args.TeamID,
		Users:    userList,
		HashPW:   string(hashed),
		Balance:  DefaultStartAmount,
		Holdings: make(map[string]string),
	}

	data := &datatypes.DataArgs{Team: team}

	// Attempt to CreateTeam and return propogated status
	status, err := ss.ls.Transact(datatypes.CreateTeam, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil
}

// Join the specified teamID if it exists and the password given is correct
func (ss *stockServer) JoinTeam(args *stockrpc.JoinTeamArgs, reply *stockrpc.JoinTeamReply) error {
	// retrieve userID from session
	userID, err := ss.RetrieveSession(args.SessionKey)
	if err != nil {
		reply.Status = datatypes.NoSuchSession
		return nil
	}

	// create argument for transaction JoinTeam
	user := datatypes.User{UserID: userID}
	team := datatypes.Team{TeamID: args.TeamID}

	data := &datatypes.DataArgs{
		User: user,
		Team: team,
		Pw:   args.Password,
	}

	// attempt to perform transaction JoinTeam, propogate status reply
	status, err := ss.ls.Transact(datatypes.JoinTeam, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil
}

// Leave the team with the specified teamID
func (ss *stockServer) LeaveTeam(args *stockrpc.LeaveTeamArgs, reply *stockrpc.LeaveTeamReply) error {
	// retrieve userID from session
	userID, err := ss.RetrieveSession(args.SessionKey)
	if err != nil {
		reply.Status = datatypes.NoSuchSession
		return nil
	}

	// create args for LeaveTeam Transaction
	user := datatypes.User{UserID: userID}
	team := datatypes.Team{TeamID: args.TeamID}

	data := &datatypes.DataArgs{
		User: user,
		Team: team,
	}

	// attempt to remove user from team
	status, err := ss.ls.Transact(storagerpc.LeaveTeam, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil
}

func (ss *stockServer) MakeTransaction(args *stockrpc.MakeTransactionArgs, reply *stockrpc.MakeTransactionReply) error {
	// retrieve userID from session

	log.Println("Got transaction request", args.Requests)



	userID, err := ss.RetrieveSession(args.SessionKey)
	if err != nil {
		reply.Status = datatypes.NoSuchSession
		return nil
	}

	user := datatypes.User{UserID: userID}

	data := &datatypes.DataArgs{
		User:     user,
		Requests: args.Requests,
	}

	status, err := ss.ls.Transact(storagerpc.MakeTransaction, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil
}

func (ss *stockServer) GetPortfolio(args *stockrpc.GetPortfolioArgs, reply *stockrpc.GetPortfolioReply) error {
	// check if team exists
	teamKey := util.CreateTeamKey(args.TeamID)
	encodedTeam, status, err := ss.ls.Get(teamKey)
	if err != nil {
		reply.Status = datatypes.BadData
		return err
	} else if status != storagerpc.OK {
		reply.Status = datatypes.NoSuchTeam
		return nil
	}

	var team datatypes.Team
	err = json.Unmarshal([]byte(encodedTeam), &team)
	if err != nil {
		reply.Status = datatypes.BadData
		return err
	}

	// get holdings from IDs
	var holding datatypes.Holding
	holdings := make([]datatypes.Holding, 0, len(team.Holdings))
	for _, holdingKey := range team.Holdings {
		holdingData, status, err := ss.ls.Get(holdingKey)
		if err != nil {
			reply.Status = datatypes.BadData
			return err
		} else if status != storagerpc.OK {
			reply.Status = datatypes.NoSuchHolding
			return nil
		}

		err = json.Unmarshal([]byte(holdingData), &holding)
		if err != nil {
			reply.Status = datatypes.BadData
			return err
		}

		holdings = append(holdings, holding)
	}

	reply.Stocks = holdings
	reply.Status = datatypes.OK

	return nil
}

func (ss *stockServer) GetPrice(args *stockrpc.GetPriceArgs, reply *stockrpc.GetPriceReply) error {
	tickerKey := util.CreateTickerKey(args.Ticker)
	tickerData, status, err := ss.ls.Get(tickerKey)
	if err != nil {
		reply.Status = datatypes.BadData
		return err
	} else if status != storagerpc.OK {
		reply.Status = datatypes.NoSuchTicker
		return nil
	}

	var ticker datatypes.Ticker
	err = json.Unmarshal([]byte(tickerData), &ticker)
	if err != nil {
		reply.Status = datatypes.BadData
		return err
	}

	reply.Price = ticker.Price
	reply.Status = datatypes.OK

	return nil
}

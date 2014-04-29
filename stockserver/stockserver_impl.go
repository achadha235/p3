package stockserver

import (
	"achadha235/p3/datatypes"
	"achadha235/p3/libstore"
	"achadha235/p3/rpc/storagerpc"
	"achadha235/p3/util"
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const (
	DefaultStartAmount = 10000000 // starting amount (in cents)
	MaxNumberTeams     = 3        // maximum number of teams a user can be on
	MaxNumberUsers     = 1000     // max number of users on a team
	MaxNumberHoldings  = 1000     // max number of holdings a user can have
)

type stockServer struct {
	server     *rpc.Server
	ls         libstore.Libstore
	sessionMap map[[]byte]string // map from sessionKey to userID for that session
}

// NewStockServer creates, starts, and returns a new StockServer.
// masterHostPort is the coordinator's host:port and myHostPort is the
// port that the StockServer should listen on for StockClient calls

func NewStockServer(masterHostPort, myHostPort string) (stockServer, error) {
	s := rpc.NewServer()

	ls, err := libstore.NewLibstore(masterHostPort, myHostPort)
	if err != nil {
		return nil, err
	}

	sMap := make(map[[]byte]string)
	ss := &stockServer{
		server:     s,
		ls:         ls,
		sessionMap: sMap,
	}

	listener, err := net.Listen("tcp", myHostPort)
	if err != nil {
		return nil, err
	}

	err = rpc.RegisterName("StockServer", stockrpc.Wrap(s))
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
	if userID, ok := ss.sessionMap[userID]; !ok {
		return "", errors.New("No such sessionKey")
	}

	return userID, nil
}

// LoginUser checks that the user exists and returns a unique session key for that user
func (ss *stockServer) LoginUser(args *LoginUserArgs, reply *LoginUserReply) error {
	key := util.CreateUserKey(args.UserID)
	// check if the user exists
	encodedUser, err := ss.ls.Get(key)
	if err != nil {
		reply.Status = datatypes.NoSuchUser
		return nil
	}

	var user datatypes.User
	err = json.Unmarshal(encodedUser, user)
	if err != nil {
		return err
	}

	// check if user pw is correct
	err = bcrypt.CompareHashAndPassword(user.hashPW, args.Password)
	if err != nil {
		reply.Status = datatypes.PermissionDenied
		return nil
	}

	// create session key
	sessionID := args.UserID + time.Now().String()
	sessionKey := bcrypt.GenerateFromPassword(sessionID, bcrypt.DefaultCost)

	// save the session key for current user
	ss.sessionMap[sessionKey] = args.UserID

	reply.SessionKey = sessionKey
	reply.Status = datatypes.OK
	return nil
}

// CreateUser adds a user to the game, or returns Exists if userID is already in use
func (ss *stockServer) CreateUser(args *CreateUserArgs, reply *CreateUserReply) error {

	hashed := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)

	// create user object
	user := datatypes.User{
		userID: args.UserID,
		hashPW: hashed,
		teams:  make([]string, 0, MaxNumberTeams),
	}

	data := &datatypes.DataArgs{user: user}

	status, err = ss.ls.Transact(datatypes.CreateUser, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil
}

// CreateTeam adds a team to the game, or returns Exists if teamID is already in use
func (ss *stockServer) CreateTeam(args *CreateTeamArgs, reply *CreateTeamReply) error {
	// lookup userID based on session
	userID, err := ss.RetrieveSession(args.SessionKey)
	if err != nil {
		reply.Status = datatypes.NoSuchSession
		return nil
	}

	// Add user to the team he created
	userList := make([]string, 0, MaxNumberUsers)
	userList = append(userList, userID)

	// Create team pw
	hashed := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)

	team := datatypes.Team{
		teamID:   args.TeamID,
		users:    userList,
		hashPW:   hashed,
		balance:  DefaultStartAmount,
		holdings: make([]datatypes.Holding, 0, MaxNumberHoldings),
	}

	data := &datatypes.DataArgs{team: team}

	// Attempt to CreateTeam and return propogated status
	status, err = ss.ls.Transact(datatypes.CreateTeam, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil
}

// Join the specified teamID if it exists and the password given is correct
func (ss *stockServer) JoinTeam(args *JoinTeamArgs, reply *JoinTeamReply) error {
	// retrieve userID from session
	userID, err := ss.RetrieveSession(args.SessionKey)
	if err != nil {
		reply.Status = datatypes.NoSuchSession
		return nil
	}

	// create argument for transaction JoinTeam
	user := datatypes.User{userID: userID}
	team := datatypes.User{teamID: args.TeamID}

	data := &datatypes.DataArgs{
		user: user,
		team: team,
		pw:   args.Password,
	}

	// attempt to perform transaction JoinTeam, propogate status reply
	status, err = ss.ls.Transact(datatypes.JoinTeam, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil
}

// Leave the team with the specified teamID
func (ss *stockServer) LeaveTeam(args *LeaveTeamArgs, reply *LeaveTeamReply) error {
	// retrieve userID from session
	userID, err := ss.RetrieveSession(args.SessionKey)
	if err != nil {
		reply.Status = datatypes.NoSuchSession
		return nil
	}

	// create args for LeaveTeam Transaction
	user := datatypes.User{userID: userID}
	team := datatypes.Team{teamID: args.TeamID}

	data := &datatypes.DataArgs{
		user: user,
		team: team,
	}

	// attempt to remove user from team
	status, err = ss.ls.Transact(storagerpc.LeaveTeam, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil
}

func (ss *stockServer) MakeTransaction(args *MakeTransactionArgs, reply *MakeTransactionReply) error {
	// retrieve userID from session
	userID, err := ss.RetrieveSession(args.SessionKey)
	if err != nil {
		reply.Status = datatypes.NoSuchSession
		return nil
	}

	user := datatypes.User{userID: userID}

	data := &datatypes.DataArgs{
		user:     user,
		requests: args.Requests,
	}

	status, err := ss.ls.Transact(storagerpc.MakeTransaction, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil
}

func (ss *stockServer) GetPortfolio(args *GetPortfolioArgs, reply *GetPortfolioReply) error {
	// check if team exists
	teamKey := util.CreateTeamKey(args.TeamID)
	encodedTeam, err := ss.ls.Get(teamKey)
	if err != nil {
		reply.Status = stockrpc.NoSuchTeam
		return nil
	}

	var team stockrpc.Team
	err = json.Unmarshal(encodedTeam, &team)
	if err != nil {
		return err
	}

	// get holdings from IDs
	var holding stockrpc.Holding
	holdings := make([]stockrpc.Holding, 0, len(hList))
	for i := 0; i < len(team.holdings); i++ {
		holdingData, err := ss.ls.Get(team.holdings[i])
		if err != nil {
			log.Println("Holding with ID not found: ", err)
			return err
		}

		err = json.Unmarshal(holdingData, &holding)
		if err != nil {
			log.Println("Holding cannot be unmarshaled")
			return err
		}

		append(holdings, holding)
	}

	reply.Stocks = holdings
	reply.Status = stockrpc.OK

	return nil
}

func (ss *stockServer) GetPrice(args *GetPriceArgs, reply *GetPriceReply) error {
	tickerKey := util.CreateTickerKey(args.Ticker)
	tickerData, err := ss.ls.Get(tickerKey)
	if err != nil {
		reply.Status = datatypes.NoSuchTicker
		return nil
	}

	var ticker datatypes.Ticker
	err = json.Unmarshal(tickerData, &ticker)
	if err != nil {
		log.Println("Unable to unmarshal Ticker: ", err)
		return err
	}

	reply.Price = ticker.price
	reply.Status = datatypes.OK

	return nil
}

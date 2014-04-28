package stockserver

import (
	"achadha235/p3/libstore"
	"achadha235/p3/rpc/stockrpc"
	"achadha235/p3/rpc/storagerpc"
	"achadha235/p3/util"
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"
)

const (
	DefaultStartAmount = 10000000 // starting amount (in cents)
	MaxNumberTeams     = 3        // maximum number of teams a user can be on
	MaxNumberUsers     = 100      // max number of users on a team
	MaxNumberHoldings  = 1000     // max number of holdings a user can have
)

type stockServer struct {
	server     *rpc.Server
	nodeID     int
	ls         libstore.Libstore
	sessionMap map[[]byte]string // map from sessionKey to userID for that session
}

// NewStockServer creates, starts, and returns a new StockServer.
// masterHostPort is the coordinator's host:port and myHostPort is the
// port that the StockServer should listen on for StockClient calls

func NewStockServer(masterHostPort, myHostPort string, nodeID int) (stockServer, error) {
	s := rpc.NewServer()

	ls, err := libstore.NewLibstore(masterHostPort, myHostPort, nodeID)
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
		reply.Status = stockrpc.NoSuchUser
		return nil
	}

	var user stockrpc.User
	err = json.Unmarshal(encodedUser, user)
	if err != nil {
		return err
	}

	// check if user pw is correct
	err = bcrypt.CompareHashAndPassword(user.hashPW, args.Password)
	if err != nil {
		reply.Status = stockrpc.PermissionDenied
		return nil
	}

	// create session key
	sessionID := args.UserID + time.Now().String()
	sessionKey := bcrypt.GenerateFromPassword(sessionID, bcrypt.DefaultCost)

	// save the session key for current user
	ss.sessionMap[sessionKey] = args.UserID

	reply.SessionKey = sessionKey
	reply.Status = stockrpc.OK
	return nil
}

// CreateUser adds a user to the game, or returns Exists if userID is already in use
func (ss *stockServer) CreateUser(args *CreateUserArgs, reply *CreateUserReply) error {

	hashed := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)

	// create user object
	user := &stockrpc.User{
		userID: args.UserID,
		hashPW: hashed,
		teams:  make([]string, 0, MaxNumberTeams),
	}

	encodedUser, err := json.Marshal(user)
	if err != nil {
		return err
	}

	status, err = ss.ls.Transact(storagerpc.CreateUser, encodedUser)
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
		reply.Status = stockrpc.NoSuchSession
		return nil
	}

	userList := make([]string, 0, MaxNumberUsers)
	userList = append(userList, userID)

	// create team pw
	hashed := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)

	team := &stockrpc.Team{
		teamID:   args.TeamID,
		users:    userList,
		hashPW:   hashed,
		balance:  DefaultStartAmount,
		holdings: make([]stockrpc.Holding, 0, MaxNumberHoldings),
	}

	encodedTeam, err := json.Marshal(team)
	if err != nil {
		return err
	}

	// Attempt to CreateTeam and return propogated status
	status, err = ss.ls.Transact(storagerpc.CreateTeam, encodedTeam)
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
		reply.Status = stockrpc.NoSuchSession
		return nil
	}

	// create argument for transaction JoinTeam
	userArgs := &stockrpc.UserTeamData{userID: userID, teamID: args.TeamID}
	data, err := json.Marshal(userArgs)
	if err != nil {
		return err
	}

	// attempt to perform transaction JoinTeam, propogate status reply
	status, err = ss.ls.Transact(storagerpc.JoinTeam, data)
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
		reply.Status = stockrpc.NoSuchSession
		return nil
	}

	// create args for LeaveTeam Transaction
	args := &stockrpc.UserTeamData{userID: userID, teamID: args.TeamID}
	data, err := json.Marshal(args)
	if err != nil {
		return err
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
		reply.Status = stockrpc.NoSuchSession
		return nil
	}

	data, err := json.Marshal(args.Requests)
	if err != nil {
		return err
	}

	status, err := ss.ls.Transact(storagerpc.MakeTransaction, data)
	if err != nil {
		return err
	}

	reply.Status = status
	return nil

	/* THIS IS THE LOGIC FOR MAKING A BUY/SELL TRANSACTION /*


	   /*	// check if team exists
	   	teamKey := util.CreateTeamKey(args.TeamID)
	   	teamList, err = ss.ls.GetList(teamKey)
	   	if err != nil {
	   		reply.Status = stockrpc.NoSuchTeam
	   		return nil
	   	}

	   	// check if ticker exists/get current price
	   	tickerKey := util.CreateTickerKey(args.Ticker)
	   	priceStr, err := ss.ls.Get(tickerKey)
	   	if err != nil {
	   		reply.Status = stockrpc.NoSuchTicker
	   		return nil
	   	}

	   	sharePrice := strconv.ParseUint(priceStr, 10, 64)
	   	// check if user is on team
	   	for i := 0; i < len(teamList); i++ {
	   		if teamList[i] == userID {
	   			break
	   		}

	   		// if user not found, then deny permission
	   		if i == len(teamList)-1 {
	   			reply.Status = stockrpc.PermissionDenied
	   			return nil
	   		}

	   		continue
	   	}

	   	// get team balance and perform update
	   	balanceKey := util.CreateTeamBalanceKey(args.TeamID)
	   	balance, err = ss.ls.GetList(balanceKey)
	   	if err != nil {
	   		reply.Status = stockrpc.PermissionDenied
	   		return nil
	   	}

	   	// get quantity of current holding
	   	holdingKey := util.CreateTeamHoldingKey(args.TeamID, args.Ticker)
	   	holdObj, err := ss.ls.Get(holdingKey)

	   	var holding stockrpc.Holding

	   	if err != nil {
	   		// if not found, then team has 0 shares
	   		amt := 0
	   	} else {
	   		err = json.Unmarshal(data, &holding)
	   		if err != nil {
	   			return err
	   		}

	   		amt := holding.quantity
	   	}

	   	newBalance := 0 // placeholder
	   	if args.Action == "buy" {
	   		// BUY TRANSACTION
	   		cost := args.Quantity * sharePrice
	   		if balance-cost < 0 {
	   			reply.Status = stockrpc.InsufficientQuantity
	   			return nil
	   		}

	   		amt += args.Quantity
	   		newBalance = balance - cost
	   	} else if args.Action == "sell" {
	   		// SELL TRANSACTION
	   		profit := args.Quantity * sharePrice
	   		// not enough shares to sell
	   		if amt < args.Quantity {
	   			reply.Status = stockrpc.InsufficientQuantity
	   			return nil
	   		}

	   		newBalance = balance + profit
	   		amt -= args.Quantity
	   	}

	   	// update balance
	   	err = ss.ls.Put(balanceKey, strconv.FormatUint(newBalance, 10))
	   	if err != nil {
	   		reply.Status = stockrpc.PermissionDenied
	   		return nil
	   	}

	   	// update holding
	   	updatedHolding := &stockrpc.Holding{ticker: args.Ticker, quantity: amt, acquired: time.Now()}
	   	jsonHolding, err := json.Marshal(updatedHolding)
	   	if err != nil {
	   		return err
	   	}

	   	err = ss.ls.Put(holdingKey, jsonHolding)
	   	if err != nil {
	   		reply.Status = stockrpc.PermissionDenied
	   		return nil
	   	}

	   	reply.Status = stockrpc.OK
	   	return nil*/
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
		reply.Status = stockrpc.NoSuchTicker
		return nil
	}

	var ticker stockrpc.Ticker
	err = json.Unmarshal(tickerData, &ticker)
	if err != nil {
		log.Println("Unable to unmarshal Ticker: ", err)
		return err
	}

	reply.Price = ticker.price
	reply.Status = stockrpc.OK

	return nil
}

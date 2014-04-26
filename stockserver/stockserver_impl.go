package stockserver

import (
	"achadha235/p3/libstore"
	"achadha235/p3/rpc/stockrpc"
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
	ss := &stockServer{server: s, ls: ls, sessionMap: sMap}

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
	userID, err := ss.ls.Get(sessionKey)
	if err != nil {
		return "", err
	}

	return userID, nil
}

// LoginUser checks that the user exists and returns a unique session key for that user
func (ss *stockServer) LoginUser(args *LoginUserArgs, reply *LoginUserReply) error {
	key := util.CreateUserKey(args.UserID)
	// check if the user exists
	_, err := ss.ls.Get(key)
	if err != nil {
		reply.Status = stockrpc.NoSuchUser
		return nil
	}

	// check if user pw is correct
	pKey := util.CreateUserPassKey
	hashPW, err := ss.ls.Get(pKey)
	if err != nil {
		reply.Status = stockrpc.NoSuchUser
		log.Println("Passkey not found on StorageServer")
		return nil
	}

	err = bcrypt.CompareHashAndPassword(hashPW, args.Password)
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
	key := util.CreateUserKey(args.UserID)
	// check if the userID exists
	_, err := ss.ls.Get(key)
	if err != nil {
		reply.Status = stockrpc.Exists
		return nil
	}

	// create user if does not exist
	err = ss.Put(key, true)
	if err != nil {
		return err
	}

	// generate pass key with bcrypt package
	pKey := util.CreateUserPassKey(args.UserID)
	hashed := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)
	err = ss.Put(pKey, string(hashed))
	if err != nil {
		return err
	}

	reply.Status = stockrpc.OK
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

	// check if the teamID exists
	teamKey := util.CreateTeamKey(args.TeamID)
	_, err = ss.ls.Get(teamKey)
	if err != nil {
		reply.Status = stockrpc.Exists
		return nil
	}

	// create team if does not exist
	err = ss.ls.AppendToList(teamKey, userID)
	if err != nil {
		reply.Status = stockrpc.PermissionDenied
		return nil
	}

	// create team balance
	balanceKey := util.CreateTeamBalanceKey(args.TeamID)
	err = ss.ls.Put(balanceKey, strconv.FormatUint(DefaultStartAmount, 10))
	if err != nil {
		reply.Status = stockrpc.PermissionDenied
		return nil
	}

	// create team pw
	hashPW := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)
	pKey := util.CreateTeamPassKey(args.TeamID)
	err = ss.ls.Put(pKey, string(hashPW))
	if err != nil {
		reply.Status = stockrpc.PermissionDenied
		return nil
	}

	reply.Status = stockrpc.OK
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

	// check if team exists
	teamKey := util.CreateTeamKey(args.TeamID)
	_, err = ss.ls.GetList(teamKey)
	if err != nil {
		reply.Status = stockrpc.NoSuchTeam
		return nil
	}

	// get hashed PW for team
	pKey := util.CreateTeamPassKey(args.TeamID)
	hashPW, err := ss.ls.Get(pKey)
	if err != nil {
		reply.Status = stockrpc.NoSuchTeam
		log.Println("No PW exists for team but team exists")
		return nil
	}

	// verify that passwords match
	err = bcrypt.CompareHashAndPassword([]byte(hashPW), []byte(args.Password))
	if err != nil {
		reply.Status = stockrpc.PermissionDenied
		return nil
	}

	// add user to team
	err = ss.ls.AppendToList(teamKey, userID)
	if err != nil {
		reply.Status = stockrpc.PermissionDenied
		return nil
	}

	reply.Status = stockrpc.OK
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

	// check if team exists
	teamKey := util.CreateTeamKey(args.TeamID)
	_, err = ss.ls.GetList(teamKey)
	if err != nil {
		reply.Status = stockrpc.NoSuchTeam
		return nil
	}

	// attempt to remove user from team
	err = ss.ls.RemoveFromList(teamKey, userID)
	if err != nil {
		reply.Status = stockrpc.NoSuchUser
		return nil
	}

	reply.Status = stockrpc.OK
	return nil
}

func (ss *stockServer) MakeTransaction(args *MakeTransactionArgs, reply *MakeTransactionReply) error {
	// retrieve userID from session
	userID, err := ss.RetrieveSession(args.SessionKey)
	if err != nil {
		reply.Status = stockrpc.NoSuchSession
		return nil
	}

	// check if team exists
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
	return nil
}

func (ss *stockServer) GetPortfolio(args *GetPortfolioArgs, reply *GetPortfolioReply) error {
	// check if team exists
	teamKey := util.CreateTeamKey(args.TeamID)
	_, err := ss.ls.GetList(teamKey)
	if err != nil {
		reply.Status = stockrpc.NoSuchTeam
		return nil
	}

	holdingKey := util.CreateTeamHoldingsKey(args.TeamID)
	hList, err := ss.ls.GetList(holdingKey)
	if err != nil {
		reply.Status = stockrpc.PermissionDenied
		return nil
	}

	holdings := make([]stockrpc.Holding, 0, len(hList))
	for i := 0; i < len(hList); i++ {
		holding, err := ss.ls.Get(hList[i])
		if err != nil {
			log.Println("Holding in hList not found")
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
	priceStr, err := ss.ls.Get(tickerKey)
	if err != nil {
		reply.Status = stockrpc.NoSuchTicker
		return nil
	}

	reply.Price = strconv.ParseUint(priceStr, 10, 64)
	reply.Status = stockrpc.OK

	return nil
}

// A runner for the client to connect to the StockServer
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/achadha235/p3/datatypes"
	"github.com/achadha235/p3/stockclient"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var port = flag.Int("port", 9010, "StockServer PortNum")
var sessionKey = []byte("")

type cmdInfo struct {
	cmdline  string
	funcname string
	nargs    int
}

type argType int

// used for parsing a transaction's arguments
const (
	TeamID argType = iota
	Ticker
	Action
	Quantity
)

func init() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Start Usage: -port=<portnum>")
		fmt.Fprintln(os.Stderr, "The crunner program is a testing tool that that creates and runs an instance")
		fmt.Fprintln(os.Stderr, "of the StockClient. Used to test StockServer calls.")
		fmt.Fprintln(os.Stderr, "After calling login, the client receives a sessionKey that must be used for")
		fmt.Fprintln(os.Stderr, "authentication for further requests involving the user's identity")
		fmt.Fprintln(os.Stderr, "Usage:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Possible commands:")
		fmt.Fprintln(os.Stderr, "  LoginUser:           login userID password")
		fmt.Fprintln(os.Stderr, "  CreateUser:          cu userID password")
		fmt.Fprintln(os.Stderr, "  CreateTeam:          ct teamID password")
		fmt.Fprintln(os.Stderr, "  JoinTeam:            jt teamID password")
		fmt.Fprintln(os.Stderr, "  LeaveTeam:           lt teamID")
		fmt.Fprintln(os.Stderr, "  MakeTransaction:     tx [teamID ticker action quantity ...]")
		fmt.Fprintln(os.Stderr, "  GetPortfolio:        gp teamID")
		fmt.Fprintln(os.Stderr, "  GetPrice:            pr ticker")
	}
}

func main() {
	flag.Parse()
	if flag.NFlag() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client, err := stockclient.NewStockClient(net.JoinHostPort("localhost", strconv.Itoa(*port)))
	if err != nil {
		log.Fatalln("Failed to create StockClient:", err)
	}

	cmdlist := []cmdInfo{
		{"login", "StockServer.LoginUser", 2},
		{"cu", "StockServer.CreateUser", 2},
		{"ct", "StockServer.CreateTeam", 2},
		{"jt", "StockServer.JoinTeam", 2},
		{"lt", "StockServer.LeaveTeam", 1},
		{"tx", "StockServer.MakeTransaction", 1},
		{"gp", "StockServer.GetPortfolio", 1},
		{"pr", "StockServer.GetPrice", 1},
	}

	cmdmap := make(map[string]cmdInfo)
	for _, j := range cmdlist {
		cmdmap[j.cmdline] = j
	}

	flag.Usage()
	input := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Please enter a command: ")
		line, err := input.ReadString('\n')
		/*		_, err := fmt.Fscanf(os.Stdin, "%s %s %s \n", &args[0], &args[1], &args[2])*/
		if err != nil {
			fmt.Println("Error Parsing Args: ", err)
			flag.Usage()
			os.Exit(1)
		}

		args := strings.Split(line, " ")
		cmd := args[0]
		lastBytes := []byte(args[len(args)-1])
		args[len(args)-1] = string(lastBytes[0 : len(lastBytes)-1])

		ci, found := cmdmap[cmd]
		if !found {
			fmt.Printf("Command not found: %s\n", cmd)
			flag.Usage()
			//os.Exit(1)
		}

		if len(args) < (ci.nargs + 1) {
			fmt.Printf("Expected args: %d, but got: %d\n", ci.nargs, len(args)-1)
			flag.Usage()
			continue
		}

		switch cmd {
		case "login": // login
			status, key, err := client.LoginUser(args[1], args[2])
			sessionKey = key
			printStatus(ci.funcname, status, err)
		case "cu": // create user
			log.Println("Attempting to create user")
			status, err := client.CreateUser(args[1], args[2])
			printStatus(ci.funcname, status, err)
		case "ct": // create team
			checkSession()
			if len(args) < 3 {
				fmt.Println("Error parsing requests")
				break
			}
			status, err := client.CreateTeam(sessionKey, args[1], args[2])
			printStatus(ci.funcname, status, err)
		case "jt": // join team

			checkSession()
			if len(args) < 3 {
				fmt.Println("Error parsing requests")
				break
			}
			status, err := client.JoinTeam(sessionKey, args[1], args[2]) // error handling for
			printStatus(ci.funcname, status, err)
		case "lt": // leave team
			checkSession()
			status, err := client.LeaveTeam(sessionKey, args[1])
			printStatus(ci.funcname, status, err)
		case "tx": // make transaction
			checkSession()
			reqs, err := parseRequests(args[1:])
			if err != nil {
				fmt.Println("Error parsing requests")
				break
			}
			status, err := client.MakeTransaction(sessionKey, reqs)
			printStatus(ci.funcname, status, err)
		case "gp":
			holdings, status, err := client.GetPortfolio(args[1])
			if status == datatypes.OK {
				printHoldings(holdings)
			} else {
				printStatus(ci.funcname, status, err)
			}
		case "pr":
			price, status, err := client.GetPrice(args[1])
			if status == datatypes.OK {
				printTicker(args[1], price)
			} else {
				printStatus(ci.funcname, status, err)
			}

		}

	} // infinite for loop reading input
}

func statusToString(status datatypes.Status) (s string) {
	switch status {
	case datatypes.OK:
		s = "OK"
	case datatypes.NoSuchUser:
		s = "NoSuchUser"
	case datatypes.NoSuchTeam:
		s = "NoSuchTeam"
	case datatypes.NoSuchTicker:
		s = "NoSuchTicker"
	case datatypes.NoSuchAction:
		s = "NoSuchAction"
	case datatypes.NoSuchSession:
		s = "NoSuchSession"
	case datatypes.InsufficientQuantity:
		s = "InsufficientQuantity"
	case datatypes.Exists:
		s = "Exists"
	case datatypes.PermissionDenied:
		s = "PermissionDenied"
	case datatypes.BadData:
		s = "BadData"
	}

	return
}

func printStatus(cmdName string, status datatypes.Status, err error) {
	if err != nil {
		fmt.Println("ERROR:", cmdName, "got error:", err)
	} else if status != datatypes.OK {
		fmt.Println(cmdName, "ERROR:", cmdName, "replied with status", statusToString(status))
	} else {
		fmt.Println(cmdName, "OK")
	}
}

func printHolding(h datatypes.Holding) {
	fmt.Printf("%s - %d - %s\n", h.Ticker, h.Quantity, h.Acquired)
}

func printTicker(ticker string, price uint64) {
	dollarStr := centsToDollarStr(price)

	fmt.Printf("%s - %s\n", ticker, dollarStr)
}

func printHoldings(holdings []datatypes.Holding) {
	for _, h := range holdings {
		printHolding(h)
	}
}

func parseRequests(argList []string) ([]datatypes.Request, error) {
	/*	argList := strings.Split(req, " ")*/
	requests := make([]datatypes.Request, 0)

	// Every request uses 4 arguments, space-delimited. Check total # of args
	if len(argList)%4 != 0 {
		fmt.Println(len(argList))
		fmt.Println("ERROR Invalid Number of Arguments for Requests: ", argList)
		return requests, errors.New("Invalid args")
	}

	var team, ticker, action string
	var quantity int64
	for i := 0; i < len(argList); i++ {
		fmt.Println("args", argList[i])
		arg := argType(i)
		switch arg {
		case TeamID:
			team = argList[i]
		case Ticker:
			ticker = argList[i]
		case Action:
			action = argList[i]
		case Quantity:
			qty, err := strconv.ParseInt(argList[i], 10, 64)
			if err != nil {
				fmt.Println("Error parsing quantity as number: ", err)
				return requests, errors.New("Invalid quantity")
			}
			quantity = qty
		}
	}

	if team == "" || ticker == "" || action == "" || quantity == 0 {
		fmt.Println("Unable to parse arguments correctly")
		return requests, errors.New("Invalid args")
	}
	// add request to return list
	request := datatypes.Request{
		TeamID:   team,
		Ticker:   ticker,
		Action:   action,
		Quantity: uint64(quantity),
	}

	fmt.Println("Trying to make request", request)
	requests = append(requests, request)

	return requests, nil
}

// checks if there is some session key set
func checkSession() {
	if sessionKey == nil {
		fmt.Println("Please login")
		flag.Usage()
	}
}

// convert uint64 cents to a dollar amount string
func centsToDollarStr(price uint64) string {
	dollars := strconv.FormatUint(price/100, 10)
	cents := strconv.FormatUint(price%100, 10)

	return "$" + dollars + "." + cents
}

// Runner for the stock server

package main

import (
	"achadha235/p3/stockserver"
	"flag"
	"log"
	"net"
	"strconv"
)

var port = flag.Int("port", 9010, "port number to listen on")

func init() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatalln("Usage: strunner <master storage server host:port>")
	}

	// Create and start the StockServer.
	hostPort := net.JoinHostPort("localhost", strconv.Itoa(*port))
	_, err := stockserver.NewStockServer(flag.Arg(0), hostPort)
	if err != nil {
		log.Fatalln("Server could not be created:", err)
	}

	// Run the StockServer forever.
	select {}
}

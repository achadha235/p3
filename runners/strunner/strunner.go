// Runner for the stock server

package main

import (
	"flag"
	"github.com/achadha235/p3/stockserver"
	"log"
)

var (
	self   = flag.String("self", "localhost:9010", "port number to listen on")
	master = flag.String("master", "localhost:9009", "hostport of masterServer")
)

func init() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
}

func main() {
	flag.Parse()
	if flag.NFlag() < 2 {
		log.Fatalln("Usage: strunner <master storage server host:port>")
	}

	// Create and start the StockServer.
	_, err := stockserver.NewStockServer(*master, *self)
	if err != nil {
		log.Fatalln("Server could not be created:", err)
	}

	// Run the StockServer forever.
	select {}
}

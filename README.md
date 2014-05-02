#WarrenBuffetGame
================

### A stock market simulator built in Go.

## Architecture
Our p3 has three main components: the client (StockClient), the game server (StockServer), and the storage server (CohortServer)
The StockServer consists of multiple layers. The top layer receives RPC calls from the client and passes the command to its Libstore.
The Libstore then performs some basic checks whether the user/team exists, and then passes the command to the coordinator (also on the game server).

The Coordinator initiates and oversees the 2PC process for the requested transaction.
We changed our design to move the Coordinator onto the StockServer in order allow for multiple coordinators, particularly for the case in which the coordinator dies, as in standard 2PC, progress can no longer be made.

## Usage

We have created runners for each of the components of our game.
These can be run as follows:
##### Storage:`./srunner -master=<masterHostPort> -self=<selfHostPort>`

Note: Additional flags can also be specified such as -N (number of storage servers in the ring) and -id (the nodeID assigned to that server)

##### GameServer:`./strunner -master=<storageMasterHostPort> -self=<selfHostPort>`

##### Client:`./crunner -port=<GameServer Port>`

In addition, we have created a wrapper around these structure for convenience. There is a shell runner in the top-level directory named `game.sh`. If you execute `chmod 7777 game.sh` (to allow permissions), you will then be able to run the client (and all other necessary components) through the shell script.

## Testing

The test file is located in `tests/` and is called `test_basic.sh`
In order to run the test file, fix permissions same as above and execute
`./test_basic.sh`

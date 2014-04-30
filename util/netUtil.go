package util

import (
	"fmt"
	"github.com/achadha235/p3/rpc/storagerpc"
	"hash/fnv"
	"net/rpc"
	"strings"
)

const (
	MaxConnectAttempts = 5
)

// return the appropriate node to use for the RPC request to the hashing ring
func FindServerFromKey(key string, servers []storagerpc.Node) *storagerpc.Node {
	if servers == nil || len(servers) == 0 {
		return nil
	} else if len(servers) == 1 {
		return &servers[0]
	}

	parts := strings.Split(key, "-")
	if len(parts) >= 2 {
		key = parts[1]
	}

	fmt.Println("key: ", key)
	hashed := StoreHash(key)

	current := servers[0]
	for i := 0; i < len(servers); i++ {
		current := servers[i]
		if hashed <= current.NodeId {
			return &current
		}
	}
	// current should be last
	if hashed > current.NodeId {
		return &servers[0]
	}

	// not found
	return nil
}

// StoreHash hashes a string key and returns a 32-bit integer.
func StoreHash(key string) uint32 {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return hasher.Sum32()
}

func TryDial(hostport string) (*rpc.Client, error) {
	maxTries := MaxConnectAttempts
	for ; ; maxTries-- {
		client, err := rpc.DialHTTP("tcp", hostport)
		if err == nil {
			return client, nil
		} else if maxTries == 0 {
			return nil, err
		} else {
			continue
		}
	}
}

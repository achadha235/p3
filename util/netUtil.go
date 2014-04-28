package util

import (
	"net/rpc"
)

const (
	MaxConnectAttempts = 5
)

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

#!/bin/bash

if [ -z $GOPATH ]; then
    echo "FAIL: GOPATH environment variable is not set"
    exit 1
fi

if [ -n "$(go version | grep 'darwin/amd64')" ]; then
    GOOS="darwin_amd64"
elif [ -n "$(go version | grep 'linux/amd64')" ]; then
    GOOS="linux_amd64"
else
    echo "FAIL: only 64-bit Mac OS X and Linux operating systems are supported"
    exit 1
fi

# stock server runner
go install github.com/achadha235/p3/runners/strunner
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# storageserver runner
go install github.com/achadha235/p3/runners/srunner
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# tester runner
go install github.com/achadha235/p3/tests/tester
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# Pick random ports between [10000, 20000).
STORAGE_PORT=$(((RANDOM % 10000) + 10000))
STOCK_PORT=$(((RANDOM % 10000) + 10000))
STORAGE_SERVER=$GOPATH/bin/srunner
STOCK_SERVER=$GOPATH/bin/strunner
TESTER = $GOPATH/bin/tester

# Start an instance of the storage server.
${STORAGE_SERVER} -port=${STORAGE_PORT} &
STORAGE_SERVER_PID=$!

# Start an instance of the stock server.
${STOCK_SERVER} -port=${STOCK_PORT} &
STOCK_SERVER_PID=$!

sleep 5

# Start the test.
${TESTER} -host="localhost:${STOCK_PORT}" "localhost:${STORAGE_PORT}"

# Kill the stock server
kill -9 ${STOCK_SERVER_PID}
wait ${STOCK_SERVER_PID}

# Kill the storage server.
kill -9 ${STORAGE_SERVER_PID}
wait ${STORAGE_SERVER_PID}

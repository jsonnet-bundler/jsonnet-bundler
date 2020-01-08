#!/usr/bin/env bash

BINARY_NAME=jb
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

HELP_FILE=$PWD/_output/help.txt
echo "$ $BINARY_NAME -h" > $HELP_FILE
$PWD/_output/$GOOS/$GOARCH/$BINARY_NAME 2>> $HELP_FILE
exit 0

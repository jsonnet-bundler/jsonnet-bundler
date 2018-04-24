#!/usr/bin/env bash

BINARY_NAME=jb

HELP_FILE=$PWD/_output/help.txt
echo "$ $BINARY_NAME -h" > $HELP_FILE
PATH=$PATH:$PWD/_output/linux/amd64 $BINARY_NAME -h 2>> $HELP_FILE
exit 0

#!/bin/sh

echo "hello"

#*****************************************************************
#************************ Building binary ******************
#*****************************************************************

go version
# echo $GOCACHE
# export GOCACHE=cache
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go

zip stripe-pdf.zip main

aws lambda update-function-code \
    --function-name  stripe-pdf \
    --zip-file fileb://./stripe-pdf.zip


#!/bin/bash

export GOOS=$1
export GOARCH=$2

#export CC=amd64-linux-gnueabihf-gcc

#MODE=c-shared
MODE=plugin

go build -buildmode=$MODE -o ../../build/$GOOS/$GOARCH/awsdig/plugins/aws/ami.plugin ./ami/ami.go 
go build -buildmode=$MODE -o ../../build/$GOOS/$GOARCH/awsdig/plugins/aws/iam.plugin ./iam/iam.go 
go build -buildmode=$MODE -o ../../build/$GOOS/$GOARCH/awsdig/plugins/aws/emr.plugin ./emr/emr.go 

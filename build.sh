#!/bin/bash

export GOOS=$1
export GOARCH=$2

MODE=plugin

mkdir -p ./build/$GOOS/$GOARCH/plugins/aws

go build -buildmode=$MODE -o ./build/$GOOS/$GOARCH/plugins/aws/ami.plugin ./aws/ami/ami.go 
go build -buildmode=$MODE -o ./build/$GOOS/$GOARCH/plugins/aws/iam.plugin ./aws/iam/iam.go 
go build -buildmode=$MODE -o ./build/$GOOS/$GOARCH/plugins/aws/emr.plugin ./aws/emr/emr.go 

#!/bin/bash

export GOOS=$1
export GOARCH=$2

MODE=plugin

mkdir -p ./build/$GOOS/$GOARCH/plugins/aws

go build -buildmode=$MODE -ldflags="-s -w" -o ./build/$GOOS/$GOARCH/plugins/aws/ami.plugin ./aws/ami/ami.go 
go build -buildmode=$MODE -ldflags="-s -w" -o ./build/$GOOS/$GOARCH/plugins/aws/emr.plugin ./aws/emr/emr.go 
go build -buildmode=$MODE -ldflags="-s -w" -o ./build/$GOOS/$GOARCH/plugins/aws/iam.plugin ./aws/iam/iam.go 
go build -buildmode=$MODE -ldflags="-s -w" -o ./build/$GOOS/$GOARCH/plugins/aws/cloudformation.plugin ./aws/cloudformation/cloudformation.go 
go build -buildmode=$MODE -ldflags="-s -w" -o ./build/$GOOS/$GOARCH/plugins/aws/route53.plugin ./aws/route53/route53.go 

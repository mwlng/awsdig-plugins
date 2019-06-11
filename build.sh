#!/bin/bash

export GOOS=$1
export GOARCH=$2

GO_PATH=`go env | grep GOPATH | cut -f 2 -d \"`

PROJ_PATH=awsdig-plugins

MODE=plugin

BUILD_PATH="${GO_PATH}/src/${PROJ_PATH}/build"

mkdir -p $BUILD_PATH/$GOOS/$GOARCH/plugins/aws

go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/ami.plugin ./aws/ami/ami.go 
go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/emr.plugin ./aws/emr/emr.go 
go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/ec2-instances.plugin ./aws/ec2/ec2-instances.go 
go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/iam.plugin ./aws/iam/iam.go 
go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/cloudformation.plugin ./aws/cloudformation/cloudformation.go 
go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/route53.plugin ./aws/route53/route53.go 
go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/autoscaling.plugin ./aws/asg/autoscaling.go 
go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/glue.plugin ./aws/glue/glue.go
go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/ecr.plugin ./aws/ecr/ecr.go
go build -buildmode=$MODE -ldflags="-s -w" -o $BUILD_PATH/$GOOS/$GOARCH/plugins/aws/ecs.plugin ./aws/ecs/ecs.go 

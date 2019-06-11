# awsdig-plugins

## Overview

This repository will be used to keep all related codes for awsdig plugins.

## Directory layout

    .
    ├── Gopkg.lock
    ├── Gopkg.toml
    ├── README.md
    ├── aws
    │   ├── ami
    │   ├── asg
    │   ├── cloudformation
    │   ├── ec2
    │   ├── ecr
    │   ├── ecs
    │   ├── emr
    │   ├── glue
    │   ├── iam
    │   └── route53
    ├── pkg
    │   ├── cache
    │   └── utils
    └── vendor

The common shared codes should be put into the pkg folder.

The individual plugin's code should be put into it's own folder within the cloud provider's directory(ie: aws).

## Plugin's interface

Below interface should be implemented in your plugin.

    type Plugin interface {
         Initialize(sess *session.Session)
         IsResourcePath(path string) bool
         GetResourcePrefixSuggestions(resourcePrefixPath string) []prompt.Suggest
         GetResourceSuggestions(resourcePath string) []prompt.Suggest
         GetResourceDetails(resourcePath string, resourceName string) interface{}
    }

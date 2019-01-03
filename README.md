# awsdig-plugins

## Overview

This repository will be used to keep all related codes for awsdig plugins.

## Directory layout

    .
    ├── README.md
    ├── aws
    │   ├── ami
    │   │   └── ami.go
    │   ├── emr
    │   │   └── emr.go
    │   └── iam
    │       └── iam.go
    ├── build.sh
    ├── cache.go
    ├── plugin.go
    ├── utils.go
    ├── ...

The common shared codes should be put into the top level directory.

The individual plugin's code should be put into it's own folder within the cloud provider's directory(ie: aws).

Important: Changing codes in the top level directory may require to re-compile/rebuild the awsdig.

## Plugin's interface

Below interface should be implemented in your plugin.

    type ServicePlugin interface {
         Initialize(sess *session.Session)
         IsResourcePath(path *string) bool
         GetResourcePrefixSuggestions(resourcePrefixPath *string) *[]prompt.Suggest
         GetResourceSuggestions(resourcePath *string) *[]prompt.Suggest
         GetResourceDetails(resourcePath *string, resourceName *string) interface{}
    }

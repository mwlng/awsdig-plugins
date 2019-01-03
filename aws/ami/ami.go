package main

import (
    "time"

    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/aws/aws-sdk-go/aws/session"

    "github.com/c-bata/go-prompt"
    "github.com/mwlng/aws-go-clients/clients"

    "awsdig/plugins"
)

var Service AMIService = AMIService{ client: nil }

var resourcePrefixSuggestions = []prompt.Suggest{}

var resourcePrefixSuggestionsMap = map[string][]prompt.Suggest{
    "/": resourcePrefixSuggestions, 
}

type AMIService struct {
    client *clients.EC2Client
    cache *plugins.Cache
}

func (s *AMIService) Initialize(sess *session.Session) {
    s.client = clients.NewClient("ec2", sess).(*clients.EC2Client)
    s.cache = plugins.NewCache(10*time.Second)
}

func (s *AMIService) IsResourcePath(path *string) bool {
    if *path == "/" { return true }
    if _, ok := resourcePrefixSuggestionsMap[*path]; ok {
        return true
    }
    return false
}

func (s *AMIService) listResourcesByPath(path string) *ec2.DescribeImagesOutput {
    return s.client.ListAMIsByOwner("self")
}

func (s *AMIService) fetchResourceList(path string) {
    if !s.cache.ShouldFetch(path) {
            return
    }
    s.cache.UpdateLastFetchedAt(path)
    ret := s.listResourcesByPath(path)
    s.cache.Store(path, ret)
    return
}

func (s *AMIService) GetResourcePrefixSuggestions(resourcePrefixPath string) *[]prompt.Suggest {
    suggestions := resourcePrefixSuggestionsMap[resourcePrefixPath]
    return &suggestions
}

func (s *AMIService) GetResourceSuggestions(resourcePath string) *[]prompt.Suggest {
    go s.fetchResourceList(resourcePath)
    x := s.cache.Load(resourcePath)
    if x == nil {
        return &[]prompt.Suggest{}
    }
    images := x.(*ec2.DescribeImagesOutput).Images
    if len(images) == 0 {
        return &[]prompt.Suggest{}
    }
    suggestions := make([]prompt.Suggest, len(images))
    for i := range images {
        suggestions[i] = prompt.Suggest {
            Text: *images[i].Name,
        }
    }
    return &suggestions
}

func (s *AMIService) GetResourceDetails(resourcePath string, resourceName string) interface{} {
    output := s.cache.Load(resourcePath)
    if output != nil {
        images := output.(*ec2.DescribeImagesOutput).Images
        for _, img := range images {
            if *img.Name == resourceName {
                return img
            }
        }
    }
    return nil    
}

package main

import (
    "fmt"
    "time"

    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/aws/aws-sdk-go/aws/session"

    "github.com/c-bata/go-prompt"
    "github.com/mwlng/aws-go-clients/clients"

    "github.com/mwlng/awsdig-plugins"
)

var resourcePrefixSuggestions = []prompt.Suggest{}

var resourcePrefixSuggestionsMap = map[string][]prompt.Suggest{
    "/": resourcePrefixSuggestions, 
}

type EC2Service struct {
    client *clients.EC2Client
    cache *plugins.Cache
}

var PluginService EC2Service

func (s *EC2Service) Initialize(sess *session.Session) {
    s.client = clients.NewClient("ec2", sess).(*clients.EC2Client)
    s.cache = plugins.NewCache(10*time.Second)
}

func (s *EC2Service) IsResourcePath(path *string) bool {
    if *path == "/" { return true }
    if _, ok := resourcePrefixSuggestionsMap[*path]; ok {
        return true
    }
    return false
}

func (s *EC2Service) GetResourcePrefixSuggestions(resourcePrefixPath *string) *[]prompt.Suggest {
    suggestions := resourcePrefixSuggestionsMap[*resourcePrefixPath]
    return &suggestions
}

func (s *EC2Service) listResourcesByPath(path string) *[]*ec2.Instance {
    return s.client.ListAllInstances()
}

func (s *EC2Service) fetchResourceList(path string) {
    if !s.cache.ShouldFetch(path) {
            return
    }
    s.cache.UpdateLastFetchedAt(path)
    ret := s.listResourcesByPath(path)
    if ret != nil {
        s.cache.Store(path, ret)
    }
    return
}

func (s *EC2Service) GetResourceSuggestions(resourcePath *string) *[]prompt.Suggest {
    go s.fetchResourceList(*resourcePath)
    x := s.cache.Load(*resourcePath)
    count := 0
    for {
        if x == nil {
            time.Sleep(100 * time.Millisecond)
            x = s.cache.Load(*resourcePath)
            count++
            if count <= 10 {
                continue
            } else {
               return &[]prompt.Suggest{}
            }
        }
        break
    }
    instances := *x.(*[]*ec2.Instance)
    if len(instances) == 0 {
        return &[]prompt.Suggest{}
    }
    suggestions := make([]prompt.Suggest, len(instances))
    for i := range instances {
        instNameId := fmt.Sprintf("%s(%s)", *instances[i].InstanceId, *instances[i].PrivateIpAddress) 
        name := GetNameFromTags(instances[i].Tags)
        if name {
            instNameId := fmt.Sprintf("%s(%s)", name, *instances[i].PrivateIpAddress)
        }
        suggestions[i] = prompt.Suggest {
            Text: instNameId,
        }
    }
    return &suggestions
}

func (s *EC2Service) GetResourceDetails(resourcePath *string, resourceName *string) interface{} {
    output := s.cache.Load(*resourcePath)
    if output != nil { 
        instances := *output.(*[]*ec2.Instance)
        for _, i := range instances {
            instNameId := fmt.Sprintf("%s(%s)", *i.InstanceId, *i.PrivateIpAddress)
            name := GetNameFromTags(i.Tags)
            if name {
                instNameId = fmt.Sprintf("%s(%s)", name, *i.PrivateIpAddress)
            }
            if instNameId  == *resourceName {
                return i
            }
        }
    }
    return nil    
}

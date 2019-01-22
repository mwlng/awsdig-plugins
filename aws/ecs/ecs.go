package main

import (
    "time"
    "strings"

    "github.com/aws/aws-sdk-go/service/ecs"
    "github.com/aws/aws-sdk-go/aws/session"

    "github.com/c-bata/go-prompt"
    "github.com/mwlng/aws-go-clients/clients"

    "github.com/mwlng/awsdig-plugins"
)

var resourcePrefixSuggestions = []prompt.Suggest{}

var resourcePrefixSuggestionsMap = map[string][]prompt.Suggest{
    "/": resourcePrefixSuggestions, 
}

var resourceTypeIndexMap = map[int]string{
   1: "cluster",
   2: "service",
   3: "task",
}

type ECSService struct {
    client *clients.ECSClient
    cache *plugins.Cache
}

var PluginService ECSService

func (s *ECSService) Initialize(sess *session.Session) {
    s.client = clients.NewClient("ecs", sess).(*clients.ECSClient)
    s.cache = plugins.NewCache(10*time.Second)
}

func (s *ECSService) IsResourcePath(inputPath *string) bool {
    if *inputPath == "/" { return true }

    if _, ok := resourcePrefixSuggestionsMap[*inputPath]; ok {
        return true
    }

    suggestions := s.GetResourceSuggestions(inputPath)
    if len(*suggestions) > 0 { return true }

    return false
}

func (s *ECSService) GetResourcePrefixSuggestions(resourcePrefixPath *string) *[]prompt.Suggest {
    suggestions := resourcePrefixSuggestionsMap[*resourcePrefixPath]
    return &suggestions
}

func (s *ECSService) listResourcesByPath(resourcePath *string) interface{}{
    if *resourcePath == "/" {
        return s.client.ListAllClusters()
    }
    pathComponents := strings.Split(*resourcePath, "/")
    pathLength := len(pathComponents)
    switch resourceTypeIndexMap[pathLength-1] {
    case "cluster":
         clusterName := pathComponents[pathLength-1]
         return s.client.ListAllServices(&clusterName)
    case "service":
        clusterName := pathComponents[pathLength-2]
        serviceName := pathComponents[pathLength-1]
        return s.client.ListAllTasks(&clusterName, &serviceName)
    //case "task":
    //    taskName := pathComponents[pathLength-1]
    //    return s.client.ListAllTaskDefinitions(&taskName)
    }
    return nil
}

func (s *ECSService) fetchResourceList(resourcePath string) {
    if !s.cache.ShouldFetch(resourcePath) {
            return
    }
    s.cache.UpdateLastFetchedAt(resourcePath)
    ret := s.listResourcesByPath(&resourcePath)
    if ret != nil {
        s.cache.Store(resourcePath, ret)
    }
    return
}

func resourcesToSuggestions(resources interface{}) *[]prompt.Suggest{
    switch resources.(type) {
    case *[]*ecs.Cluster:
         l := len(*resources.(*[]*ecs.Cluster))
         if l != 0 {
             suggestions := make([]prompt.Suggest, l)
             for i, r := range *resources.(*[]*ecs.Cluster) {
                 suggestions[i] = prompt.Suggest {
                     Text: *r.ClusterName,
                 }
             }
             return &suggestions
         }
    case *[]*ecs.Service:
        l := len(*resources.(*[]*ecs.Service))
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*ecs.Service) {
                suggestions[i] = prompt.Suggest { 
                    Text: *r.ServiceName, 
                }
            }
            return &suggestions
        }
    case *[]*ecs.Task:
        l := len(*resources.(*[]*ecs.Task))
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*ecs.Task) {
                suggestions[i] = prompt.Suggest { 
                    Text: *r.TaskArn, 
                }
            }
            return &suggestions
        }
    }
    return &[]prompt.Suggest{}
}

func (s *ECSService) GetResourceSuggestions(resourcePath *string) *[]prompt.Suggest {
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
    return resourcesToSuggestions(x)
}

func (s *ECSService) GetResourceDetails(resourcePath *string, resourceName *string) interface{} {
    output := s.cache.Load(*resourcePath)
    if output != nil {
        switch output.(type) {
        case *[]*ecs.Cluster: 
            clusters := *output.(*[]*ecs.Cluster)
            for _, c := range clusters {
                name := *c.ClusterName 
                if name == *resourceName {
                   return c
                }
            }
        case *[]*ecs.Service:
            services := *output.(*[]*ecs.Service)
            for _, s := range services {
                name := *s.ServiceName
                if name == *resourceName {
                   return s
                }
            }
        }
    }
    return nil    
}

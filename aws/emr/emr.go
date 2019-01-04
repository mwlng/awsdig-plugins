package main

import (
    "fmt"
    "time"

    "github.com/aws/aws-sdk-go/service/emr"
    "github.com/aws/aws-sdk-go/aws/session"

    "github.com/c-bata/go-prompt"
    "github.com/mwlng/aws-go-clients/clients"

    "awsdig/plugins"
)

var resourcePrefixSuggestions = []prompt.Suggest{}

var resourcePrefixSuggestionsMap = map[string][]prompt.Suggest{
    "/": resourcePrefixSuggestions, 
}

type EMRService struct {
    client *clients.EMRClient
    cache *plugins.Cache
}

var PluginService EMRService

func (s *EMRService) Initialize(sess *session.Session) {
    s.client = clients.NewClient("emr", sess).(*clients.EMRClient)
    s.cache = plugins.NewCache(10*time.Second)
}

func (s *EMRService) IsResourcePath(path *string) bool {
    if *path == "/" { return true }
    if _, ok := resourcePrefixSuggestionsMap[*path]; ok {
        return true
    }
    return false
}

func (s *EMRService) GetResourcePrefixSuggestions(resourcePrefixPath *string) *[]prompt.Suggest {
    suggestions := resourcePrefixSuggestionsMap[*resourcePrefixPath]
    return &suggestions
}

func (s *EMRService) listResourcesByPath(path string) *[]*emr.ClusterSummary {
    states := []string{
        "STARTING",
        "BOOTSTRAPPING",
        "RUNNING",
        "WAITING",
        "TERMINATING",
    }
    clusterStates := make([]*string, len(states))
    for i := range states {
        clusterStates[i] = &states[i]
    }
    return s.client.ListClusters(clusterStates)
}

func (s *EMRService) fetchResourceList(path string) {
    if !s.cache.ShouldFetch(path) {
            return
    }
    s.cache.UpdateLastFetchedAt(path)
    ret := s.listResourcesByPath(path)
    s.cache.Store(path, ret)
    return
}

func (s *EMRService) GetResourceSuggestions(resourcePath *string) *[]prompt.Suggest {
    go s.fetchResourceList(*resourcePath)
    x := s.cache.Load(*resourcePath)
    if x == nil {
        return &[]prompt.Suggest{}
    }
    clusters := *x.(*[]*emr.ClusterSummary)
    if len(clusters) == 0 {
        return &[]prompt.Suggest{}
    }
    suggestions := make([]prompt.Suggest, len(clusters))
    for i := range clusters {
        suggestions[i] = prompt.Suggest {
            Text: fmt.Sprintf("%s(%s)", *clusters[i].Name, *clusters[i].Id),
        }
    }
    return &suggestions
}

func (s *EMRService) GetResourceDetails(resourcePath *string, resourceName *string) interface{} {
    output := s.cache.Load(*resourcePath)
    if output != nil { 
        clusters := *output.(*[]*emr.ClusterSummary)
        for _, clus := range clusters {
            clusterId := *clus.Id
            clusterNameId := fmt.Sprintf("%s(%s)", *clus.Name, *clus.Id)
            if clusterNameId  == *resourceName {
                return s.client.DescribeCluster(&clusterId)
                //return clus
            }
        }
    }
    return nil    
}

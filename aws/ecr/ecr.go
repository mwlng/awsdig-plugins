package main

import (
    "fmt"
    "time"
    "path"
    "strings"

    "github.com/aws/aws-sdk-go/service/ecr"
    "github.com/aws/aws-sdk-go/aws/session"

    "github.com/c-bata/go-prompt"
    "github.com/mwlng/aws-go-clients/clients"

    "github.com/mwlng/awsdig-plugins"
)

var resourcePrefixSuggestions = []prompt.Suggest{
}

var resourcePrefixSuggestionsMap = map[string][]prompt.Suggest {
    "/": []prompt.Suggest{},
}

type ECRService struct {
    client *clients.ECRClient
    cache *plugins.Cache
}

var PluginService ECRService

func (s *ECRService) Initialize(sess *session.Session) {
    s.client = clients.NewClient("ecr", sess).(*clients.ECRClient)
    s.cache = plugins.NewCache(10*time.Second)
}

func (s *ECRService) IsResourcePath(inputPath *string) bool {
    if _, ok := resourcePrefixSuggestionsMap[*inputPath]; ok {
        return true
    } else {
        dir := path.Dir(*inputPath)
        if dir == "/" {
            return true
        } 
    }
    return false
} 

func (s *ECRService) listResourcesByPath(resourcePath string) interface{} {
    dir := path.Dir(resourcePath)
    _, base := path.Split(resourcePath)
    if dir == "/" {
        return s.client.ListRepositories()
    } else {
       repoName := strings.Replace(base, "\\/", "/", -1)
       return s.client.ListImageIdsByRepository(&repoName)
    }
    return nil
}

func (s *ECRService) fetchResourceList(resourcePath string) {
    if !s.cache.ShouldFetch(resourcePath) {
            return
    }
    s.cache.UpdateLastFetchedAt(resourcePath)
    ret := s.listResourcesByPath(resourcePath)
    if ret != nil {
        s.cache.Store(resourcePath, ret)
    }
    return
}

func resourcesToSuggestions(resources interface{}) *[]prompt.Suggest{
    switch resources.(type) {
    case *[]*ecr.Repository:
        l := len(*resources.(*[]*ecr.Repository)) 
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*ecr.Repository) {
                suggestions[i] = prompt.Suggest {
                    Text: strings.Replace(*r.RepositoryName, "/", "\\/", -1),
                }
            }
            return &suggestions
        }
    case *[]*ecr.ImageIdentifier:
        l := len(*resources.(*[]*ecr.ImageIdentifier)) 
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*ecr.ImageIdentifier) {
                suggestions[i] = prompt.Suggest {
                    Text: *r.ImageTag,
                }
            }
            return &suggestions
        }
    } 
    return &[]prompt.Suggest{}
}

func (s *ECRService) GetResourcePrefixSuggestions(resourcePrefixPath *string) *[]prompt.Suggest {
    suggestions := resourcePrefixSuggestionsMap[*resourcePrefixPath]
    return &suggestions
}

func (s *ECRService) GetResourceSuggestions(resourcePath *string) *[]prompt.Suggest {
    if *resourcePath == "/" {
        suggestions := resourcePrefixSuggestionsMap[*resourcePath]
        if len(suggestions) != 0 {
            return &suggestions
        }
    }
    go s.fetchResourceList(*resourcePath)
    x := s.cache.Load(*resourcePath)
    if x == nil {
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
    }
    switch x.(type) {
    case *[]*ecr.Repository:
      repositories:= x.(*[]*ecr.Repository)
      return resourcesToSuggestions(repositories)
    case *[]*ecr.ImageIdentifier:
      imageIds := x.(*[]*ecr.ImageIdentifier)
      return resourcesToSuggestions(imageIds)
    }
    return &[]prompt.Suggest{}
}

func (s *ECRService) GetResourceDetails(resourcePath *string, resourceName *string) interface{} {
    output := s.cache.Load(*resourcePath)
    if output != nil {
        switch output.(type) {
        case *[]*ecr.Repository:
            for _, r := range *output.(*[]*ecr.Repository) {
                fmt.Println(*resourceName)
                if *resourceName == strings.Replace(*r.RepositoryName, "/", "\\/", -1) { return r }
            }
        case *[]*ecr.ImageIdentifier:
            for _, i := range *output.(*[]*ecr.ImageIdentifier) {
                if *resourceName == *i.ImageTag {
                    return s.client.DescribeImageById(i)   
                }
            }
        }
    } 
    return nil    
}

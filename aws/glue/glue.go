package main

import (
    "time"
    "path"

    "github.com/aws/aws-sdk-go/service/glue"
    "github.com/aws/aws-sdk-go/aws/session"

    "github.com/c-bata/go-prompt"
    "github.com/mwlng/aws-go-clients/clients"

    "github.com/mwlng/awsdig-plugins"
)

var resourcePrefixSuggestions = []prompt.Suggest{
    {"databases", "Glue databases"},
    {"crawlers", "Glue crawlers"},
    {"classifiers",  "Glue classifiers"},
    {"triggers", "Glue job triggers"},
}

var resourcePrefixSuggestionsMap = map[string][]prompt.Suggest {
    "/" : resourcePrefixSuggestions,
    "/databases": []prompt.Suggest{},
    "/crawlers": []prompt.Suggest{},
    "/classifiers": []prompt.Suggest{},
    "/triggers": []prompt.Suggest{},
}

type GlueService struct {
    client *clients.GlueClient
    cache *plugins.Cache
}

var PluginService GlueService

func (s *GlueService) Initialize(sess *session.Session) {
    s.client = clients.NewClient("glue", sess).(*clients.GlueClient)
    s.cache = plugins.NewCache(10*time.Second)
}

func (s *GlueService) IsResourcePath(inputPath *string) bool {
    if _, ok := resourcePrefixSuggestionsMap[*inputPath]; ok {
        return true
    } else {
        dir := path.Dir(*inputPath)
        if dir == "/databases" {
            return true
        } 
    }
    return false
} 

func (s *GlueService) listResourcesByPath(resourcePath string) interface{} {
    dir := path.Dir(resourcePath)
    _, base := path.Split(resourcePath)
    if dir == "/" {
        switch base {
        case "databases":
            return s.client.ListDatabases()
        case "crawlers":
            return s.client.ListCrawlers()
        case "classifiers":
            return s.client.ListClassifiers()
        case "triggers":
            return s.client.ListTriggers()
        }
    } else {
        _, parent := path.Split(dir)
        //x := s.cache.Load(dir)
        switch parent {
            case "databases":
                return s.client.ListTables(&base)
        }
    }
    return nil
}

func (s *GlueService) fetchResourceList(resourcePath string) {
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
    case *[]*glue.Database:
        l := len(*resources.(*[]*glue.Database)) 
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*glue.Database) {
                suggestions[i] = prompt.Suggest {
                    Text: *r.Name,
                }
            }
            return &suggestions
        }
    case *[]*glue.Table:
        l := len(*resources.(*[]*glue.Table)) 
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*glue.Table) {
                suggestions[i] = prompt.Suggest {
                    Text: *r.Name,
                }
            }
            return &suggestions
        }
    case *[]*glue.Crawler:
        l := len(*resources.(*[]*glue.Crawler))
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*glue.Crawler) {
                suggestions[i] = prompt.Suggest {
                    Text: *r.Name,
                }
            }
            return &suggestions
        }
    case *[]*glue.Classifier:
        l := len(*resources.(*[]*glue.Classifier))
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, _ := range *resources.(*[]*glue.Classifier) {
                suggestions[i] = prompt.Suggest {
                    Text: "*r.Name",
                }
            }
            return &suggestions
        }
    case *[]*glue.Trigger:
        l := len(*resources.(*[]*glue.Trigger))
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*glue.Trigger) {
                suggestions[i] = prompt.Suggest {
                    Text: *r.Name,
                }
            }
            return &suggestions
        }
    } 
    return &[]prompt.Suggest{}
}

func (s *GlueService) GetResourcePrefixSuggestions(resourcePrefixPath *string) *[]prompt.Suggest {
    suggestions := resourcePrefixSuggestionsMap[*resourcePrefixPath]
    return &suggestions
}

func (s *GlueService) GetResourceSuggestions(resourcePath *string) *[]prompt.Suggest {
    if *resourcePath == "/" {
        suggestions := resourcePrefixSuggestionsMap[*resourcePath]
        return &suggestions
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
    case *[]*glue.Database:
      databases := x.(*[]*glue.Database)
      return resourcesToSuggestions(databases)
    case *[]*glue.Table:
      tables := x.(*[]*glue.Table)
      return resourcesToSuggestions(tables)
    case *[]*glue.Crawler:
      crawlers := x.(*[]*glue.Crawler)
      return resourcesToSuggestions(crawlers)
    case *[]*glue.Classifier:
      classifiers := x.(*[]*glue.Classifier)
      return resourcesToSuggestions(classifiers)
    case *[]*glue.Trigger:
      triggers := x.(*[]*glue.Trigger)
      return resourcesToSuggestions(triggers)
    }
    return &[]prompt.Suggest{}
}

func (s *GlueService) GetResourceDetails(resourcePath *string, resourceName *string) interface{} {
    output := s.cache.Load(*resourcePath)
    if output != nil {
        switch output.(type) {
        case *[]*glue.Database:
            for _, d := range *output.(*[]*glue.Database) {
                if *resourceName == *d.Name { return d }
            }
        case *[]*glue.Table:
            for _, t := range *output.(*[]*glue.Table) {
                if *resourceName == *t.Name { return t }
            }
        case *[]*glue.Crawler:
            for _, c := range *output.(*[]*glue.Crawler) {
                if *resourceName == *c.Name { return c }
            }
        case *[]*glue.Classifier:
            for _, c := range *output.(*[]*glue.Classifier) {
                if *resourceName == "*r.Name" { return c }
            }
        case *[]*glue.Trigger:
            for _, t := range *output.(*[]*glue.Trigger) {
                if *resourceName == *t.Name { return t } 
            }
        }
    } 
    return nil    
}

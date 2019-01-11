package main

import (
    "fmt"
    "time"
    //"strings"
    "path"
    //"path/filepath"
    //"encoding/json"

    "github.com/aws/aws-sdk-go/service/route53"
    "github.com/aws/aws-sdk-go/aws/session"

    "github.com/c-bata/go-prompt"
    "github.com/mwlng/aws-go-clients/clients"

    "github.com/mwlng/awsdig-plugins"
)

var resourcePrefixSuggestions = []prompt.Suggest{
    {"zones", "Route53 hosted zones"},
    {"geolocations", "Route53 geographic locations"},
}

var resourcePrefixSuggestionsMap = map[string][]prompt.Suggest {
    "/" : resourcePrefixSuggestions,
}

type R53Service struct {
    client *clients.R53Client
    cache *plugins.Cache
}

var PluginService R53Service

func (s *R53Service) Initialize(sess *session.Session) {
    s.client = clients.NewClient("route53", sess).(*clients.R53Client)
    s.cache = plugins.NewCache(10*time.Second)
}

func (s *R53Service) IsResourcePath(path *string) bool {
    if *path == "/" { return true }
    if _, ok := resourcePrefixSuggestionsMap[*path]; ok {
        return true
    }
    suggestions := s.GetResourceSuggestions(path)
    if len(*suggestions) > 0 { return true }
    return false
} 

func (s *R53Service) listResourcesByPath(resourcePath string) interface{} {
    _, base := path.Split(resourcePath)
    switch base {
    case "zones":
        return s.client.ListHostedZones()
    case "geolocations":
        return s.client.ListGeoLocations()
    }
    return nil
}

func (s *R53Service) fetchResourceList(resourcePath string) {
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

func extractGeoLocation(input *route53.GeoLocationDetails) *string {
    var name = "Unknown"
    if input.ContinentName != nil && input.CountryName !=nil && input.SubdivisionName != nil {
        name = fmt.Sprintf("%s-%s-%s", *input.ContinentName, *input.CountryName, *input.SubdivisionName)
    } else if input.ContinentName != nil && input.CountryName != nil {
        name = fmt.Sprintf("%s-%s", *input.ContinentName, *input.CountryName)
    } else if input.CountryName != nil && input.SubdivisionName != nil {
        name = fmt.Sprintf("%s-%s", *input.CountryName, *input.SubdivisionName)
    } else if input.ContinentName != nil {
        name = *input.ContinentName
    } else if input.CountryName != nil {
        name = *input.CountryName
    } else if input.SubdivisionName != nil {
        name = *input.SubdivisionName
    }
    return &name
}

func resourcesToSuggestions(resources interface{}) *[]prompt.Suggest{
    switch resources.(type) {
    case *[]*route53.HostedZone:
         l := len(*resources.(*[]*route53.HostedZone)) 
         if l != 0 {
             suggestions := make([]prompt.Suggest, l)
             for i, r := range *resources.(*[]*route53.HostedZone) {
                 suggestions[i] = prompt.Suggest { Text: *r.Name, }
             }
             return &suggestions
         }
    case *[]*route53.GeoLocationDetails:
        l := len(*resources.(*[]*route53.GeoLocationDetails))
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*route53.GeoLocationDetails) {
                name := extractGeoLocation(r)
                suggestions[i] = prompt.Suggest { Text: *name, }
            }
            return &suggestions
        }
    } 
    return &[]prompt.Suggest{}
}

func (s *R53Service) GetResourcePrefixSuggestions(resourcePrefixPath *string) *[]prompt.Suggest {
    suggestions := resourcePrefixSuggestionsMap[*resourcePrefixPath]
    return &suggestions
}

func (s *R53Service) GetResourceSuggestions(resourcePath *string) *[]prompt.Suggest {
    //paths := strings.Split(*resourcePath, "/")
    //realPath := fmt.Sprintf("/%s", path.Join(paths[0:2]...))
    if *resourcePath == "/" { 
        suggestions := resourcePrefixSuggestionsMap[*resourcePath]
        return &suggestions
    }
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
    _, base := path.Split(*resourcePath)
    switch base {
    case "zones":
      zones := x.(*[]*route53.HostedZone)
      return resourcesToSuggestions(zones)
    case "geolocations":
      locations := x.(*[]*route53.GeoLocationDetails)
      return resourcesToSuggestions(locations)
    }
    return &[]prompt.Suggest{}
}

func (s *R53Service) GetResourceDetails(resourcePath *string, resourceName *string) interface{} {
    output := s.cache.Load(*resourcePath)
    if output != nil {
        switch output.(type) {
        case *[]*route53.HostedZone:
            for _, z := range *output.(*[]*route53.HostedZone) {
                if *resourceName == *z.Name { return z }
            }
        case *[]*route53.GeoLocationDetails:
            for _, g := range *output.(*[]*route53.GeoLocationDetails) {
                if *resourceName == *extractGeoLocation(g) { return g }
            }
        }
    } 
     /*else {
        dir := path.Dir(*resourcePath) 
        output := s.cache.Load(dir)
        if output != nil {
            _, base := path.Split(*resourcePath)
            switch output.(type) {
            case *[]*cloudformation.StackSummary:
                for _, stack := range *output.(*[]*cloudformation.StackSummary) {
                    if base == *stack.StackName { 
                        switch *resourceName {
                        case "template":
                            return s.client.GetTemplate(&base)
                        case "resources":
                            return s.client.ListStackResources(&base)
                        case "changesets":
                            return s.client.ListChangeSets(&base)
                        }
                    }
                }
            case *[]*cloudformation.StackSetSummary:
                for _, ss := range *output.(*[]*cloudformation.StackSetSummary) {
                    if base == *ss.StackSetName { 
                        switch *resourceName {
                        case "instances":
                            //return &instances
                        }
                    }
                }
            }
        }
    } */
    return nil    
}

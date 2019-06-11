package main

import (
	"time"

	"awsdig-plugins/pkg/cache"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/c-bata/go-prompt"
	"github.com/mwlng/aws-go-clients/clients"
)

var (
	PluginService ASGService

	resourcePrefixSuggestions    = []prompt.Suggest{}
	resourcePrefixSuggestionsMap = map[string][]prompt.Suggest{
		"/": resourcePrefixSuggestions,
	}
)

type ASGService struct {
	client *clients.ASGClient
	cache  *cache.Cache
}

func (s *ASGService) Initialize(sess *session.Session) {
	s.client = clients.NewClient("autoscaling", sess).(*clients.ASGClient)
	s.cache = cache.NewCache(10 * time.Second)
}

func (s *ASGService) IsResourcePath(path string) bool {
	if path == "/" {
		return true
	}
	if _, ok := resourcePrefixSuggestionsMap[path]; ok {
		return true
	}
	return false
}

func (s *ASGService) GetResourcePrefixSuggestions(resourcePrefixPath string) []prompt.Suggest {
	suggestions := resourcePrefixSuggestionsMap[resourcePrefixPath]
	return suggestions
}

func (s *ASGService) listResourcesByPath(path string) []*autoscaling.Group {
	return s.client.ListAllAutoScalingGroups()
}

func (s *ASGService) fetchResourceList(path string) {
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

func (s *ASGService) GetResourceSuggestions(resourcePath string) []prompt.Suggest {
	go s.fetchResourceList(resourcePath)
	x := s.cache.Load(resourcePath)
	count := 0
	for {
		if x == nil {
			time.Sleep(100 * time.Millisecond)
			x = s.cache.Load(resourcePath)
			count++
			if count <= 10 {
				continue
			} else {
				return []prompt.Suggest{}
			}
		}
		break
	}
	groups := x.([]*autoscaling.Group)
	if len(groups) == 0 {
		return []prompt.Suggest{}
	}
	suggestions := make([]prompt.Suggest, len(groups))
	for i := range groups {
		suggestions[i] = prompt.Suggest{
			Text: *groups[i].AutoScalingGroupName,
		}
	}
	return suggestions
}

func (s *ASGService) GetResourceDetails(resourcePath string, resourceName string) interface{} {
	output := s.cache.Load(resourcePath)
	if output != nil {
		groups := output.([]*autoscaling.Group)
		for _, g := range groups {
			name := g.AutoScalingGroupName
			if *name == resourceName {
				return g
			}
		}
	}
	return nil
}

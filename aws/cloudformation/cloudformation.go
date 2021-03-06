package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"awsdig-plugins/pkg/cache"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/c-bata/go-prompt"
	"github.com/mwlng/aws-go-clients/clients"
)

var (
	PluginService CFNService

	resourcePrefixSuggestions = []prompt.Suggest{
		{"stacks", "Cloudformation stacks"},
		{"stacksets", "Cloudformation stacksets"},
	}
	resourcePrefixSuggestionsMap = map[string][]prompt.Suggest{
		"/":          resourcePrefixSuggestions,
		"/stacks":    []prompt.Suggest{},
		"/stacksets": []prompt.Suggest{},
	}
	stackSuggestions = []prompt.Suggest{
		{"template", "Stack's  template"},
		{"resources", "Stack's resources"},
		{"changesets", "Stack's changesets"},
	}
	stacksetSuggestions = []prompt.Suggest{
		{"instances", "Stack instances"},
	}
)

type CFNService struct {
	client *clients.CFNClient
	cache  *cache.Cache
}

func (s *CFNService) Initialize(sess *session.Session) {
	s.client = clients.NewClient("cloudformation", sess).(*clients.CFNClient)
	s.cache = cache.NewCache(10 * time.Second)
}

func (s *CFNService) IsResourcePath(inputPath string) bool {
	if inputPath == "/" {
		return true
	}
	if _, ok := resourcePrefixSuggestionsMap[inputPath]; ok {
		return true
	}
	if inputPath == "/stacks" || inputPath == "/stacksets" {
		s.GetResourceSuggestions(inputPath)
		return true
	} else {
		dir := path.Dir(inputPath)
		if dir == "/stacks" || dir == "/stacksets" {
			return true
		}
	}
	return false
}

func (s *CFNService) listResourcesByPath(resourcePath string) interface{} {
	_, base := path.Split(resourcePath)
	switch base {
	case "stacks":
		return s.client.ListStacks()
	case "stacksets":
		return s.client.ListStackSets()
	}
	return nil
}

func (s *CFNService) fetchResourceList(resourcePath string) {
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

func resourcesToSuggestions(resources interface{}) []prompt.Suggest {
	switch resources.(type) {
	case []*cloudformation.StackSummary:
		l := len(resources.([]*cloudformation.StackSummary))
		if l != 0 {
			suggestions := []prompt.Suggest{}
			for _, r := range resources.([]*cloudformation.StackSummary) {
				if *r.StackStatus != "DELETE_COMPLETE" {
					suggestions = append(suggestions, prompt.Suggest{
						Text: *r.StackName,
					})
				}
			}
			return suggestions
		}
	case []*cloudformation.StackSetSummary:
		l := len(resources.([]*cloudformation.StackSetSummary))
		if l != 0 {
			suggestions := make([]prompt.Suggest, l)
			for i, r := range resources.([]*cloudformation.StackSetSummary) {
				suggestions[i] = prompt.Suggest{
					Text: *r.StackSetName,
				}
			}
			return suggestions
		}
	}
	return []prompt.Suggest{}
}

func (s *CFNService) GetResourcePrefixSuggestions(resourcePrefixPath string) []prompt.Suggest {
	return resourcePrefixSuggestionsMap[resourcePrefixPath]
}

func (s *CFNService) GetResourceSuggestions(resourcePath string) []prompt.Suggest {
	if resourcePath == "/" {
		return resourcePrefixSuggestionsMap[resourcePath]
	}
	paths := strings.Split(resourcePath, "/")
	realPath := fmt.Sprintf("/%s", path.Join(paths[0:2]...))
	go s.fetchResourceList(realPath)
	x := s.cache.Load(resourcePath)
	if x == nil {
		_, base := path.Split(filepath.Dir(resourcePath))
		switch base {
		case "stacks":
			return stackSuggestions
		case "stacksets":
			return stacksetSuggestions
		}
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
	}
	_, base := path.Split(resourcePath)
	switch base {
	case "stacks":
		stacks := x.([]*cloudformation.StackSummary)
		return resourcesToSuggestions(stacks)
	case "stacksets":
		stacksets := x.([]*cloudformation.StackSetSummary)
		return resourcesToSuggestions(stacksets)
	}
	return []prompt.Suggest{}
}

func (s *CFNService) GetResourceDetails(resourcePath string, resourceName string) interface{} {
	output := s.cache.Load(resourcePath)
	if output != nil {
		switch output.(type) {
		case []*cloudformation.StackSummary:
			for _, s := range output.([]*cloudformation.StackSummary) {
				if resourceName == *s.StackName {
					return s
				}
			}
		case []*cloudformation.StackSetSummary:
			for _, ss := range output.([]*cloudformation.StackSetSummary) {
				if resourceName == *ss.StackSetName {
					return ss
				}
			}
		}
	} else {
		dir := path.Dir(resourcePath)
		output := s.cache.Load(dir)
		if output != nil {
			_, base := path.Split(resourcePath)
			switch output.(type) {
			case []*cloudformation.StackSummary:
				for _, stack := range output.([]*cloudformation.StackSummary) {
					if base == *stack.StackName {
						switch resourceName {
						case "template":
							return s.client.GetTemplate(&base)
						case "resources":
							return s.client.ListStackResources(&base)
						case "changesets":
							return s.client.ListChangeSets(&base)
						}
					}
				}
			case []*cloudformation.StackSetSummary:
				for _, ss := range output.([]*cloudformation.StackSetSummary) {
					if base == *ss.StackSetName {
						switch resourceName {
						case "instances":
							//return &instances
						}
					}
				}
			}
		}
	}
	return nil
}

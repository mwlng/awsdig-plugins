package main

import (
	"time"

	"awsdig-plugins/pkg/cache"
	"awsdig-plugins/pkg/utils"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/c-bata/go-prompt"
	"github.com/mwlng/aws-go-clients/clients"
)

var (
	PluginService ECSService

	resourcePrefixSuggestions = []prompt.Suggest{
		{"clusters", "ECS clusters"},
		{"taskdefs", "ECS task definitions"},
	}
	resourcePrefixSuggestionsMap = map[string][]prompt.Suggest{
		"/":         resourcePrefixSuggestions,
		"/clusters": []prompt.Suggest{},
		"/taskdefs": []prompt.Suggest{},
	}
	resourceTypeMap = map[int]string{
		1: "/",
		2: "cluster",
		3: "service",
		4: "task",
	}
)

type ECSService struct {
	client *clients.ECSClient
	cache  *cache.Cache
}

func (s *ECSService) Initialize(sess *session.Session) {
	s.client = clients.NewClient("ecs", sess).(*clients.ECSClient)
	s.cache = cache.NewCache(10 * time.Second)
}

func (s *ECSService) IsResourcePath(inputPath string) bool {
	if _, ok := resourcePrefixSuggestionsMap[inputPath]; ok {
		return true
	}

	if inputPath == "/clusters" || inputPath == "/taskdefs" {
		return true
	}

	pathComponents := utils.PathToStrings(inputPath)
	l := len(pathComponents)
	if (pathComponents)[1] == "clusters" {
		switch resourceTypeMap[l-1] {
		case "cluster":
			x := s.cache.Load("/clusters")
			if x != nil {
				return true
			}
		case "service":
			x := s.cache.Load(inputPath)
			if x != nil {
				return true
			}
			return true
		}
	}

	return false
}

func (s *ECSService) GetResourcePrefixSuggestions(resourcePrefixPath string) []prompt.Suggest {
	return resourcePrefixSuggestionsMap[resourcePrefixPath]
}

func (s *ECSService) listResourcesByPath(resourcePath string) interface{} {
	pathComponents := utils.PathToStrings(resourcePath)
	l := len(pathComponents)
	base := pathComponents[l-1]
	switch base {
	case "clusters":
		return s.client.ListClusters()
	case "taskdefs":
		return s.client.ListTaskDefinitions()
	}

	switch resourceTypeMap[l-1] {
	case "cluster":
		clusterName := pathComponents[l-1]
		return s.client.ListServicesByCluster(&clusterName)
	case "service":
		clusterName := pathComponents[l-2]
		serviceName := pathComponents[l-1]
		return s.client.ListTasksByService(&clusterName, &serviceName)
	}
	return nil
}

func (s *ECSService) fetchResourceList(resourcePath string) {
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
	case []*ecs.Cluster:
		l := len(resources.([]*ecs.Cluster))
		if l != 0 {
			suggestions := make([]prompt.Suggest, l)
			for i, r := range resources.([]*ecs.Cluster) {
				suggestions[i] = prompt.Suggest{
					Text: *r.ClusterName,
				}
			}
			return suggestions
		}
	case []*ecs.Service:
		l := len(resources.([]*ecs.Service))
		if l != 0 {
			suggestions := make([]prompt.Suggest, l)
			for i, r := range resources.([]*ecs.Service) {
				suggestions[i] = prompt.Suggest{
					Text: *r.ServiceName,
				}
			}
			return suggestions
		}
	case []*string:
		l := len(resources.([]*string))
		if l != 0 {
			suggestions := make([]prompt.Suggest, l)
			for i, r := range resources.([]*string) {
				suggestions[i] = prompt.Suggest{
					Text: *r,
				}
			}
			return suggestions
		}
	}
	return []prompt.Suggest{}
}

func (s *ECSService) GetResourceSuggestions(resourcePath string) []prompt.Suggest {
	if resourcePath == "/" {
		suggestions := resourcePrefixSuggestionsMap[resourcePath]
		return suggestions
	}

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
	return resourcesToSuggestions(x)
}

func (s *ECSService) GetResourceDetails(resourcePath string, resourceName string) interface{} {
	output := s.cache.Load(resourcePath)
	if output != nil {
		switch output.(type) {
		case []*ecs.Cluster:
			clusters := output.([]*ecs.Cluster)
			for _, c := range clusters {
				name := *c.ClusterName
				if name == resourceName {
					return c
				}
			}
		case []*ecs.Service:
			services := output.([]*ecs.Service)
			for _, s := range services {
				name := *s.ServiceName
				if name == resourceName {
					return s
				}
			}
		case []*string:
			return s.client.DescribeTaskDefinition(&resourceName)
		}
	}
	return nil
}

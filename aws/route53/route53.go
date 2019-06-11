package main

import (
	"fmt"
	"path"
	"regexp"
	"time"

	"awsdig-plugins/pkg/cache"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/c-bata/go-prompt"
	"github.com/mwlng/aws-go-clients/clients"
)

var (
	PluginService R53Service

	resourcePrefixSuggestions = []prompt.Suggest{
		{"zones", "Route53 hosted zones"},
		{"geolocations", "Route53 geographic locations"},
	}
	resourcePrefixSuggestionsMap = map[string][]prompt.Suggest{
		"/": resourcePrefixSuggestions,
	}
)

type R53Service struct {
	client *clients.R53Client
	cache  *cache.Cache
}

func (s *R53Service) Initialize(sess *session.Session) {
	s.client = clients.NewClient("route53", sess).(*clients.R53Client)
	s.cache = cache.NewCache(10 * time.Second)
}

func (s *R53Service) IsResourcePath(inputPath string) bool {
	if inputPath == "/" {
		return true
	}
	if _, ok := resourcePrefixSuggestionsMap[inputPath]; ok {
		return true
	}
	r, _ := regexp.Compile("/geolocations/.*")
	if r.MatchString(inputPath) {
		return false
	}
	if inputPath == "/zones" || inputPath == "/geolocations" {
		s.GetResourceSuggestions(inputPath)
		return true
	} else {
		dir := path.Dir(inputPath)
		if dir == "/zones" {
			return true
		}
	}
	return false
}

func (s *R53Service) getHostedZoneIdByName(hostedZoneName string, hostedZones []*route53.HostedZone) *string {
	for _, z := range hostedZones {
		_, id := path.Split(*z.Id)
		if fmt.Sprintf("%s(%s)", *z.Name, id) == hostedZoneName {
			return z.Id
		}
	}
	return nil
}

func (s *R53Service) listResourcesByPath(resourcePath string) interface{} {
	dir := path.Dir(resourcePath)
	_, base := path.Split(resourcePath)
	if dir == "/" {
		switch base {
		case "zones":
			return s.client.ListHostedZones()
		case "geolocations":
			return s.client.ListGeoLocations()
		}
	} else {
		_, parent := path.Split(dir)
		x := s.cache.Load(dir)
		switch parent {
		case "zones":
			id := s.getHostedZoneIdByName(base, x.([]*route53.HostedZone))
			if id != nil {
				return s.client.ListResourceRecordSets(id)
			}
		}
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

func extractGeoLocation(input *route53.GeoLocationDetails) string {
	var name = "Unknown"
	if input.ContinentName != nil && input.CountryName != nil && input.SubdivisionName != nil {
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
	return name
}

func resourcesToSuggestions(resources interface{}) []prompt.Suggest {
	switch resources.(type) {
	case []*route53.HostedZone:
		l := len(resources.([]*route53.HostedZone))
		if l != 0 {
			suggestions := make([]prompt.Suggest, l)
			for i, r := range resources.([]*route53.HostedZone) {
				_, id := path.Split(*r.Id)
				suggestions[i] = prompt.Suggest{
					Text: fmt.Sprintf("%s(%s)", *r.Name, id),
				}
			}
			return suggestions
		}
	case []*route53.GeoLocationDetails:
		l := len(resources.([]*route53.GeoLocationDetails))
		if l != 0 {
			suggestions := make([]prompt.Suggest, l)
			for i, r := range resources.([]*route53.GeoLocationDetails) {
				name := extractGeoLocation(r)
				suggestions[i] = prompt.Suggest{Text: name}
			}
			return suggestions
		}
	case []*route53.ResourceRecordSet:
		l := len(resources.([]*route53.ResourceRecordSet))
		if l != 0 {
			suggestions := make([]prompt.Suggest, l)
			for i, r := range resources.([]*route53.ResourceRecordSet) {
				name := fmt.Sprintf("%s(%s)", *r.Name, *r.Type)
				suggestions[i] = prompt.Suggest{Text: name}
			}
			return suggestions
		}
	}
	return []prompt.Suggest{}
}

func (s *R53Service) GetResourcePrefixSuggestions(resourcePrefixPath string) []prompt.Suggest {
	return resourcePrefixSuggestionsMap[resourcePrefixPath]
}

func (s *R53Service) GetResourceSuggestions(resourcePath string) []prompt.Suggest {
	//paths := strings.Split(*resourcePath, "/")
	//realPath := fmt.Sprintf("/%s", path.Join(paths[0:2]...))
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
	switch x.(type) {
	case []*route53.HostedZone:
		zones := x.([]*route53.HostedZone)
		return resourcesToSuggestions(zones)
	case []*route53.GeoLocationDetails:
		locations := x.([]*route53.GeoLocationDetails)
		return resourcesToSuggestions(locations)
	case []*route53.ResourceRecordSet:
		records := x.([]*route53.ResourceRecordSet)
		return resourcesToSuggestions(records)
	}
	return []prompt.Suggest{}
}

func (s *R53Service) GetResourceDetails(resourcePath string, resourceName string) interface{} {
	output := s.cache.Load(resourcePath)
	if output != nil {
		switch output.(type) {
		case []*route53.HostedZone:
			for _, z := range output.([]*route53.HostedZone) {
				_, id := path.Split(*z.Id)
				if resourceName == fmt.Sprintf("%s(%s)", *z.Name, id) {
					return z
				}
			}
		case []*route53.GeoLocationDetails:
			for _, g := range output.([]*route53.GeoLocationDetails) {
				if resourceName == extractGeoLocation(g) {
					return g
				}
			}
		case []*route53.ResourceRecordSet:
			for _, r := range output.([]*route53.ResourceRecordSet) {
				name := fmt.Sprintf("%s(%s)", *r.Name, *r.Type)
				if resourceName == name {
					return r
				}
			}
		}
	}
	return nil
}

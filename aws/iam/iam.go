package main

import (
    "fmt"
    "time"
    "strings"
    "path"
    "path/filepath"
    "encoding/json"

    "github.com/aws/aws-sdk-go/service/iam"
    "github.com/aws/aws-sdk-go/aws/session"

    "github.com/c-bata/go-prompt"
    "github.com/mwlng/aws-go-clients/clients"

    "github.com/mwlng/awsdig-plugins"
)

var resourcePrefixSuggestions = []prompt.Suggest{
    {"users", "IAM users"},
    {"groups", "IAM groups"},
    {"policies",  "IAM policies"},
    {"roles", "IAM role"},
}

var resourcePrefixSuggestionsMap = map[string][]prompt.Suggest {
    "/" : resourcePrefixSuggestions,
    "/users": []prompt.Suggest{},
    "/groups": []prompt.Suggest{},
    "/policies": []prompt.Suggest{},
    "/roles": []prompt.Suggest{},
}

var userSuggestions = []prompt.Suggest {
    {"inline", "User's inline IAM policies"},
    {"policies", "User's attached IAM policies"},
    {"groups", "User's IAM groups"},
}

var groupSuggestions = []prompt.Suggest {
    {"inline", "Group's inline policies"},
    {"policies", "Group's attached IAM policies"},
}

var roleSuggestions = []prompt.Suggest {
    {"inline", "Role's inline policies"},
    {"policies", "Role's attached IAM policies"},
}

var policySuggestions = []prompt.Suggest {
    {"document", "Policy's document"},
}

type IAMService struct {
    client *clients.IAMClient
    cache *plugins.Cache
}

var PluginService IAMService

func (s *IAMService) Initialize(sess *session.Session) {
    s.client = clients.NewClient("iam", sess).(*clients.IAMClient)
    s.cache = plugins.NewCache(10*time.Second)
}

func (s *IAMService) IsResourcePath(path *string) bool {
    if *path == "/" { return true }
    if _, ok := resourcePrefixSuggestionsMap[*path]; ok {
        return true
    }
    suggestions := s.GetResourceSuggestions(path)
    if len(*suggestions) > 0 { return true }
    return false
} 

func (s *IAMService) listResourcesByPath(resourcePath string) interface{} {
    _, base := path.Split(resourcePath)
    switch base {
    case "users":
        return s.client.ListUsers()
    case "groups":
        return s.client.ListGroups()
    case "roles":
        return s.client.ListRoles()
    case "policies":
        return s.client.ListPolicies()
    }
    return nil
}

func (s *IAMService) fetchResourceList(resourcePath string) {
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
    case *[]*iam.User:
        l := len(*resources.(*[]*iam.User)) 
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*iam.User) {
                suggestions[i] = prompt.Suggest {
                    Text: *r.UserName,
                }
            }
            return &suggestions
        }
    case *[]*iam.Group:
        l := len(*resources.(*[]*iam.Group))
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*iam.Group) {
                suggestions[i] = prompt.Suggest {
                    Text: *r.GroupName,
                }
            }
            return &suggestions
        }
    case *[]*iam.Role:
        l := len(*resources.(*[]*iam.Role))
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*iam.Role) {
                suggestions[i] = prompt.Suggest {
                    Text: *r.RoleName,
                }
            }
            return &suggestions
        }
    case *[]*iam.Policy:
        l := len(*resources.(*[]*iam.Policy))
        if l != 0 {
            suggestions := make([]prompt.Suggest, l)
            for i, r := range *resources.(*[]*iam.Policy) {
                suggestions[i] = prompt.Suggest {
                    Text: *r.PolicyName,
                }
            }
            return &suggestions
        }
    } 
    return &[]prompt.Suggest{}
}

func (s *IAMService) GetResourcePrefixSuggestions(resourcePrefixPath *string) *[]prompt.Suggest {
    suggestions := resourcePrefixSuggestionsMap[*resourcePrefixPath]
    return &suggestions
}

func (s *IAMService) GetResourceSuggestions(resourcePath *string) *[]prompt.Suggest {
    paths := strings.Split(*resourcePath, "/")
    realPath := fmt.Sprintf("/%s", path.Join(paths[0:2]...))
    go s.fetchResourceList(realPath)
    x := s.cache.Load(*resourcePath)
    if x == nil {
        _, base := path.Split(filepath.Dir(*resourcePath))
        switch base {
            case "users":
                return &userSuggestions
            case "groups":
                return &groupSuggestions
            case "roles":
                return &roleSuggestions
            case "policies":
                return &policySuggestions
        }
        return &[]prompt.Suggest{}
    }
    _, base := path.Split(*resourcePath)
    switch base {
    case "users":
      users := x.(*[]*iam.User)
      return resourcesToSuggestions(users)
    case "groups":
      groups := x.(*[]*iam.Group)
      return resourcesToSuggestions(groups)
    case "roles":
      roles := x.(*[]*iam.Role)
      return resourcesToSuggestions(roles)
    case "policies":
      policies := x.(*[]*iam.Policy)
      return resourcesToSuggestions(policies)
    }
    return &[]prompt.Suggest{}
}

func (s *IAMService) GetResourceDetails(resourcePath *string, resourceName *string) interface{} {
    output := s.cache.Load(*resourcePath)
    if output != nil {
        switch output.(type) {
        case *[]*iam.User:
            for _, u := range *output.(*[]*iam.User) {
                if *resourceName == *u.UserName { return u }
            }
        case *[]*iam.Group:
            for _, g := range *output.(*[]*iam.Group) {
                if *resourceName == *g.GroupName { return g }
            }
        case *[]*iam.Role:
            for _, r := range *output.(*[]*iam.Role) {
                if *resourceName == *r.RoleName { 
                    policyDocument := plugins.UrlDecode(r.AssumeRolePolicyDocument)
                    r.AssumeRolePolicyDocument = policyDocument
                    return r 
                }
            }
        case *[]*iam.Policy:
            for _, p := range *output.(*[]*iam.Policy) {
                if *resourceName == *p.PolicyName { return p } 
            }
        }
    } else {
        dir := path.Dir(*resourcePath) 
        output := s.cache.Load(dir)
        if output != nil {
            _, base := path.Split(*resourcePath)
            switch output.(type) {
            case *[]*iam.User:
                for _, u := range *output.(*[]*iam.User) {
                    if base == *u.UserName { 
                        switch *resourceName {
                        case "inline":
                            policyNames := s.client.ListUserPolicies(&base)
                            policies := make(map[string]map[string]interface{})
                            for _, p := range *policyNames {
                                var policy map[string]interface{}
                                policyDocument := s.client.GetUserPolicy(&base, p)
                                json.Unmarshal([]byte(*plugins.UrlDecode(policyDocument)), policies[*p])
                                policies[*p] = policy
                            }
                            return &policies
                        case "policies":
                            return s.client.ListAttachedUserPolicies(&base)
                        case "groups":
                            return s.client.ListGroupsForUser(&base)
                        }
                    }
                }
            case *[]*iam.Group:
                for _, g := range *output.(*[]*iam.Group) {
                    if base == *g.GroupName { 
                        switch *resourceName {
                        case "inline":
                            policyNames := s.client.ListGroupPolicies(&base)
                            policies := make(map[string]map[string]interface{})
                            for _, p := range *policyNames {
                                var policy map[string]interface{}
                                policyDocument := s.client.GetGroupPolicy(&base, p)
                                json.Unmarshal([]byte(*plugins.UrlDecode(policyDocument)), &policy)
                                policies[*p] = policy
                            }
                            return &policies
                        case "policies":
                            return s.client.ListAttachedGroupPolicies(&base)
                        }
                    }
                }
            case *[]*iam.Role:
                for _, r := range *output.(*[]*iam.Role) {
                    if base == *r.RoleName { 
                        switch *resourceName {
                        case "inline":
                            policyNames := s.client.ListRolePolicies(&base)
                            policies := make(map[string]map[string]interface{})
                            for _, p := range *policyNames {
                                var policy map[string]interface{}
                                policyDocument := s.client.GetRolePolicy(&base, p)
                                json.Unmarshal([]byte(*plugins.UrlDecode(policyDocument)), &policy)
                                policies[*p] = policy
                            }
                            return &policies
                        case "policies":
                            return s.client.ListAttachedRolePolicies(&base)
                        }
                    }
                }
            case *[]*iam.Policy:
                for _, p := range *output.(*[]*iam.Policy) {
                    if base == *p.PolicyName {
                        switch *resourceName {
                        case "document":
                            policyDocument := s.client.GetPolicyVersion(p.Arn, p.DefaultVersionId).Document
                            var policy map[string]interface{}
                            json.Unmarshal([]byte(*plugins.UrlDecode(policyDocument)), &policy)
                            return &policy
                        }
                    }
                }
            }
        }
    }
    return nil    
}

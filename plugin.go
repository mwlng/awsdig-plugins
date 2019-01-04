package plugins

import (
   "fmt"
   "log"
   "strings"
   "path"
   "path/filepath"
   "plugin"

   "github.com/aws/aws-sdk-go/aws/session"

   "github.com/c-bata/go-prompt"
)

var pluginsGlob []string = []string{"*.so", "*.plugin"}

type ServicePlugin interface {
     Initialize(sess *session.Session) 
     IsResourcePath(path *string) bool
     GetResourcePrefixSuggestions(resourcePrefixPath *string) *[]prompt.Suggest
     GetResourceSuggestions(resourcePath *string) *[]prompt.Suggest
     GetResourceDetails(resourcePath *string, resourceName *string) interface{}
}

type PluginsManager struct {
    provider string
    sess *session.Session
    pluginsDir string
    pluginsList []string
    pluginsMap map[string]ServicePlugin
}

func NewPluginsManager(provider string, 
                       pluginsDir string, 
                       sess *session.Session) *PluginsManager {
    pm := PluginsManager{
        provider: provider,
        pluginsDir: pluginsDir,
        sess: sess,
    }
    pm.initialize()
    return &pm
}

func (pm *PluginsManager) initialize() {
    var plugins []string

    pDir := path.Join(pm.pluginsDir, pm.provider)
    for _, glob := range pluginsGlob {
        files, err := filepath.Glob(fmt.Sprintf("%s/%s", pDir, glob))
        if err != nil {
            log.Fatal(err)
        }
        plugins = append(plugins, files...)
    }
        
    ret := make([]string, len(plugins))
    for i, f := range plugins {
        ret[i] = path.Base(f)
    } 
    pm.pluginsList = ret
    pm.loadPlugins()
}

func (pm *PluginsManager) loadPlugins() {
    pm.pluginsMap = map[string]ServicePlugin{}
    pDir := path.Join(pm.pluginsDir, pm.provider)
    for _, p := range pm.pluginsList {
        plug, err := plugin.Open(fmt.Sprintf("%s/%s", pDir, p))
        if err != nil {
            log.Fatal(err)
        }
        service := strings.ToLower(strings.Split(p, ".")[0])
	plugin, err := plug.Lookup("PluginService")
	if err != nil {
            log.Fatal(err)
	}
        plugin.(ServicePlugin).Initialize(pm.sess)
        pm.pluginsMap[service] = plugin.(ServicePlugin)
    }
}

func (pm *PluginsManager) GetPluginsList() *[]string {
    return &pm.pluginsList 
}

func (pm *PluginsManager) GetPluginNames() *[]string {
    ret := make([]string, len(pm.pluginsList))
    for i, p := range pm.pluginsList {
        ret[i] = strings.ToLower(strings.Split(p, ".")[0])    
    }
    return &ret
}

func (pm *PluginsManager) IsPluginFound(service string) bool {
    for _, p := range *pm.GetPluginNames() {
        if p == service { return true }
    }
    return false
}

func (pm *PluginsManager) GetServiceSuggestions() *[]prompt.Suggest {
    serviceSuggestions := []prompt.Suggest{}
    for _, service := range *pm.GetPluginNames() {
        serviceSuggestions = append(serviceSuggestions,
                                    prompt.Suggest{Text: service})
    }
    return &serviceSuggestions
}

func (pm *PluginsManager) GetServicePlugin(service string) ServicePlugin {
    if plugin, ok := pm.pluginsMap[service]; ok {
        return plugin
    }
    return nil
}

func (pm *PluginsManager) GetValidResourcePath(service *string, 
                                               resourcePath *string) *string {
     validPath := *resourcePath
     if plugin, ok := pm.pluginsMap[*service]; ok {
         for {
              if plugin.IsResourcePath(&validPath) {
                  return &validPath
              } 
              validPath = filepath.Dir(validPath)
              if validPath == "/" { break }
         }
     }
     return &validPath
}

func (pm *PluginsManager) IsResourcePath(service *string,
                                         resourcePath *string) bool {
     if plugin, ok := pm.pluginsMap[*service]; ok {
         return plugin.IsResourcePath(resourcePath)
     }
     return false
}

func (pm *PluginsManager) GetResourcePrefixSuggestions(service string, 
                                                       resourcePath string) *[]prompt.Suggest {
    if plugin, ok := pm.pluginsMap[service]; ok {
        return plugin.GetResourcePrefixSuggestions(&resourcePath)
    }
    return &[]prompt.Suggest{}
} 

func (pm *PluginsManager) GetResourceSuggestions(service string, 
                                                 resourcePath string) *[]prompt.Suggest {
    if plugin, ok := pm.pluginsMap[service]; ok {
        return plugin.GetResourceSuggestions(&resourcePath)
    }
    return &[]prompt.Suggest{} 
}

func (pm *PluginsManager) GetResourceDetails(service string, 
                                           resourcePath string, 
                                           resourceName string) interface{} {
    if plugin, ok := pm.pluginsMap[service]; ok {
        return plugin.GetResourceDetails(&resourcePath, &resourceName)
    }
    return nil    
}

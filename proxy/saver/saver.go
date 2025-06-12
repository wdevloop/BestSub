package saver

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bestruirui/bestsub/config"
	"github.com/bestruirui/bestsub/providerconfig"
	"github.com/bestruirui/bestsub/proxy/info"
	"github.com/bestruirui/bestsub/utils/log"
	"gopkg.in/yaml.v3"
)

type ProxyCategory struct {
	Name     string
	Proxies  []map[string]any
	Filter   func(result info.Proxy) bool
	Provider *providerconfig.Provider
}

type ConfigSaver struct {
	results     *[]info.Proxy
	categories  []ProxyCategory
	saveMethods []func([]byte, string) error
}

func NewConfigSaver(results *[]info.Proxy) *ConfigSaver {
	saver := &ConfigSaver{
		results:     results,
		saveMethods: chooseSaveMethods(),
		categories:  make([]ProxyCategory, 0),
	}

	providers, err := providerconfig.Load(config.GlobalConfig.ProviderFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("provider file not found: %v", config.GlobalConfig.ProviderFile)
		} else {
			log.Error("load provider file failed: %v", err)
		}
	}

	for _, p := range providers {
		prov := p
		saver.categories = append(saver.categories, ProxyCategory{
			Name:     prov.Output,
			Proxies:  make([]map[string]any, 0),
			Provider: &prov,
		})
	}

	if len(saver.categories) == 0 {
		saver.categories = defaultCategories()
	}

	return saver
}

func defaultCategories() []ProxyCategory {
	return []ProxyCategory{
		{
			Name:    "all.yaml",
			Proxies: make([]map[string]any, 0),
			Filter:  func(result info.Proxy) bool { return result.Info.Alive },
		},
		{
			Name:    "speed.yaml",
			Proxies: make([]map[string]any, 0),
			Filter:  func(result info.Proxy) bool { return result.Info.Speed > config.GlobalConfig.Check.MinSpeed },
		},
		{
			Name:    "openai.yaml",
			Proxies: make([]map[string]any, 0),
			Filter:  func(result info.Proxy) bool { return result.Info.Unlock.Chatgpt },
		},
		{
			Name:    "youtube.yaml",
			Proxies: make([]map[string]any, 0),
			Filter:  func(result info.Proxy) bool { return result.Info.Unlock.Youtube },
		},
		{
			Name:    "netflix.yaml",
			Proxies: make([]map[string]any, 0),
			Filter:  func(result info.Proxy) bool { return result.Info.Unlock.Netflix },
		},
		{
			Name:    "disney.yaml",
			Proxies: make([]map[string]any, 0),
			Filter:  func(result info.Proxy) bool { return result.Info.Unlock.Disney },
		},
	}
}

func SaveConfig(results *[]info.Proxy) {
	if len(config.GlobalConfig.Save.BeforeSaveDo) > 0 {
		if err := BeforeSaveDo(results); err != nil {
			log.Error("Failed to execute before-save scripts: %v", err)
		}
	}

	saver := NewConfigSaver(results)
	if err := saver.Save(); err != nil {
		log.Error("save config failed: %v", err)
	}

	if len(config.GlobalConfig.Save.AfterSaveDo) > 0 {
		if err := AfterSaveDo(results); err != nil {
			log.Error("Failed to execute after-save scripts: %v", err)
		}
	}
}

// SaveResults writes only the results.json file without generating providers.
func SaveResults(results *[]info.Proxy) {
	saver := NewConfigSaver(results)
	saver.saveResults()
}

// GenerateProviders categorizes proxies by provider rules and saves provider files
// without modifying or saving the results JSON.
func GenerateProviders(results *[]info.Proxy) {
	saver := NewConfigSaver(results)
	saver.categorizeProxies()
	for _, category := range saver.categories {
		if err := saver.saveCategory(category); err != nil {
			log.Error("save %s category failed: %v", category.Name, err)
		}
	}
}

// LoadResults reads a results.json file into a slice of Proxy objects.
func LoadResults(path string) ([]info.Proxy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var results []info.Proxy
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (cs *ConfigSaver) Save() error {
	cs.categorizeProxies()
	cs.saveResults()

	for _, category := range cs.categories {
		if err := cs.saveCategory(category); err != nil {
			log.Error("save %s category failed: %v", category.Name, err)
			continue
		}
	}

	return nil
}

func (cs *ConfigSaver) saveResults() {
	jsonData, err := json.MarshalIndent(cs.results, "", "  ")
	if err != nil {
		log.Error("serialize results failed: %v", err)
		return
	}
	for _, saveMethod := range cs.saveMethods {
		if err := saveMethod(jsonData, "results.json"); err != nil {
			log.Error("save results failed with one method: %v", err)
		}
	}
}

func (cs *ConfigSaver) categorizeProxies() {
	for i := range cs.categories {
		if cs.categories[i].Provider != nil {
			selected := providerconfig.Apply(*cs.categories[i].Provider, *cs.results)
			for _, p := range selected {
				cs.categories[i].Proxies = append(cs.categories[i].Proxies, p.Raw)
			}
			continue
		}
		for _, result := range *cs.results {
			if cs.categories[i].Filter != nil && cs.categories[i].Filter(result) {
				cs.categories[i].Proxies = append(cs.categories[i].Proxies, result.Raw)
			}
		}
	}
}

func (cs *ConfigSaver) saveCategory(category ProxyCategory) error {
	if len(category.Proxies) == 0 {
		log.Warn("%s proxies are empty, skip", category.Name)
		return nil
	}
	log.Debug("save %s category %v proxies", category.Name, len(category.Proxies))
	yamlData, err := yaml.Marshal(map[string]any{
		"proxies": category.Proxies,
	})
	if err != nil {
		return fmt.Errorf("serialize %s failed: %w", category.Name, err)
	}

	for _, saveMethod := range cs.saveMethods {
		if err := saveMethod(yamlData, category.Name); err != nil {
			log.Error("save %s failed with one method: %v", category.Name, err)
		}
	}

	return nil
}

func chooseSaveMethods() []func([]byte, string) error {
	methods := make([]func([]byte, string) error, 0)

	for _, methodName := range config.GlobalConfig.Save.Method {
		switch methodName {
		case "r2":
			if err := ValiR2Config(); err == nil {
				methods = append(methods, UploadToR2Storage)
			} else {
				log.Error("R2 config is incomplete: %v", err)
			}
		case "gist":
			if err := ValiGistConfig(); err == nil {
				methods = append(methods, UploadToGist)
			} else {
				log.Error("Gist config is incomplete: %v", err)
			}
		case "webdav":
			if err := ValiWebDAVConfig(); err == nil {
				methods = append(methods, UploadToWebDAV)
			} else {
				log.Error("WebDAV config is incomplete: %v", err)
			}
		case "http":
			if err := ValiHTTPConfig(); err == nil {
				methods = append(methods, SaveToHTTP)
			} else {
				log.Error("HTTP config is incomplete: %v", err)
			}
		case "local":
			methods = append(methods, SaveToLocal)
		default:
			log.Error("unknown save method: %s", methodName)
		}
	}

	if len(methods) == 0 {
		log.Warn("no valid save methods configured, using local save only")
		methods = append(methods, SaveToLocal)
	}

	return methods
}

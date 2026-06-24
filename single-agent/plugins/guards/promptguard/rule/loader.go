package rule

import (
	"os"

	"github.com/goccy/go-yaml"
)

func LoadFromFile(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var out struct {
		Rules []Rule `yaml:"rules"`
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out.Rules, nil
}

func LoadFromFiles(paths ...string) ([]Rule, error) {
	var all []Rule
	for _, p := range paths {
		rules, err := LoadFromFile(p)
		if err != nil {
			return nil, err
		}
		all = append(all, rules...)
	}
	return all, nil
}

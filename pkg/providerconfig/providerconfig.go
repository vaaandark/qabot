package providerconfig

import (
	"encoding/json"
	"os"
)

type ProviderConfig struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Model string `json:"model,omitempty"`
	Key   string `json:"key"`
}

func LoadProviderConfigFromFile(path string) ([]ProviderConfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := []ProviderConfig{}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

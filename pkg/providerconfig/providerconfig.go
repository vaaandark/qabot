package providerconfig

import (
	"encoding/json"
	"os"
	"sync/atomic"
)

type ProviderConfig struct {
	Name  string   `json:"name"`
	Url   string   `json:"url"`
	Model string   `json:"model,omitempty"`
	Keys  []string `json:"keys"`
	index uint64
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

func (pc ProviderConfig) NextKey() string {
	idx := atomic.AddUint64(&pc.index, 1) - 1
	return pc.Keys[int(idx)%len(pc.Keys)]
}

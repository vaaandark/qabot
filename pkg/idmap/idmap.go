package idmap

import (
	"encoding/json"
	"os"
)

type IdMap map[string]string

func (im IdMap) LookupName(namespacedId string) *string {
	if name, exist := im[namespacedId]; exist {
		return &name
	}
	return nil
}

type Mapping struct {
	NamespacedId string `json:"namespaced_id"`
	Name         string `json:"name"`
}

func LoadIdMapFromFile(path string) (*IdMap, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	mappings := &[]Mapping{}
	err = json.Unmarshal(bytes, mappings)
	if err != nil {
		return nil, err
	}

	idMap := make(IdMap)
	for _, mapping := range *mappings {
		idMap[mapping.NamespacedId] = mapping.Name
	}
	return &idMap, nil
}

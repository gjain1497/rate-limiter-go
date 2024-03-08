package main

import (
	"encoding/json"
	"os"
)

func readIPLateLimitingConfig(filePath string) (*IPRateLimitingMappingConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &IPRateLimitingMappingConfig{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

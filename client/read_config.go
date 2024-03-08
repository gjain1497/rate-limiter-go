package main

import (
	"encoding/json"
	"os"
)

func readIPConfig(filePath string) (*IpAddressesConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := &IpAddressesConfig{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

package cmd

import (
	"cluster-codex/internal/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Inclusion - Struct to match JSON structure
type Inclusion struct {
	Namespace string   `json:"namespace"`
	Resources []string `json:"resources"`
}

type Config struct {
	Inclusions []Inclusion `json:"inclusions"`
}

// GetNamespacesFromJSON - Function to read the JSON file and extract namespaces
func GetNamespacesFromJSON(filePath string) ([]string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if filePath == defaultFilterFileName {
			config.ClxLogger.Debug("Default filter file did not exist.")
			return nil, nil
		}
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file contents
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Parse JSON
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Extract namespaces
	var namespaces []string
	for _, inclusion := range config.Inclusions {
		namespaces = append(namespaces, inclusion.Namespace)
	}

	return namespaces, nil
}

package component

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

// DifficultyConfig represents the configuration for a difficulty level
type DifficultyConfig struct {
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// ValidateDifficulty checks if the given difficulty is valid according to the loaded configuration
func ValidateDifficulty(difficulty string) bool {
	if difficulty == "all" {
		return true
	}

	diffs, err := LoadDifficulties()
	if err != nil {
		return false
	}

	// Check against loaded difficulties (case-insensitive)
	for k := range diffs {
		if strings.EqualFold(difficulty, k) {
			return true
		}
	}
	return false
}

// LoadDifficulties loads difficulty configurations from JSON file
func LoadDifficulties() (map[string]DifficultyConfig, error) {
	data, err := ioutil.ReadFile("config/difficulties.json")
	if err != nil {
		log.Printf("Error reading difficulties.json: %v", err)
		return getDefaultDifficulties(), err
	}

	var difficulties map[string]DifficultyConfig
	if err := json.Unmarshal(data, &difficulties); err != nil {
		log.Printf("Error parsing difficulties.json: %v", err)
		return getDefaultDifficulties(), err
	}

	return difficulties, nil
}

// getDefaultDifficulties returns hardcoded default difficulties as fallback
func getDefaultDifficulties() map[string]DifficultyConfig {
	return map[string]DifficultyConfig{
		"basic": {
			Name:        "Basic",
			Icon:        "🟢",
			Color:       "#4CAF50",
			Description: "Standard rules",
		},
		"intermediate": {
			Name:        "Intermediate",
			Icon:        "🟡",
			Color:       "#FF9800",
			Description: "More challenging",
		},
		"hard": {
			Name:        "Hard",
			Icon:        "🔴",
			Color:       "#F44336",
			Description: "Expert level",
		},
		"expert": {
			Name:        "Expert",
			Icon:        "🟣",
			Color:       "#9C27B0",
			Description: "Master level",
		},
		"fun": {
			Name:        "Fun",
			Icon:        "🎉",
			Color:       "#E91E63",
			Description: "Quirky rules",
		},
	}
}

package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func GetConfig() (AssignmentConfig, error) {
	file, err := os.ReadFile(filepath.Join(".subsys", "config.json"))
	if err != nil {
		err = errors.New("this is not a subsys directory")
		return AssignmentConfig{}, err
	}

	var data AssignmentConfig
	json.Unmarshal(file, &data)
	return data, nil
}

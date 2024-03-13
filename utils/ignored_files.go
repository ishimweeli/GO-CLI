package utils

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func GetIgnoredFiles(dir string) ([]string, error) {
	ignoredFiles, err := ioutil.ReadFile(filepath.Join(dir))
	if err != nil {
		fmt.Printf("Error reading subsysignore file: %v", err)
		return []string{}, err
	}

	lines := strings.Split(string(ignoredFiles), "\n")
	return lines, nil
}

func Contains(arr []string, str string) bool {
	for _, item := range arr {
		if strings.Contains(str, item) {
			return true
		}
	}
	return false
}

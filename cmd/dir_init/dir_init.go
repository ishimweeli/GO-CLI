package dirinit

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"amalitech.org/subsys/utils"
)

type configInitializer struct {
	ProjectName string
	Directory   string
}

func NewDirectoryInitializer() (*configInitializer, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	_, project := filepath.Split(path)

	return &configInitializer{
		ProjectName: project,
		Directory:   path,
	}, nil
}

func (si *configInitializer) Initialize() error {
	fmt.Printf("Initializing a new assignment in directory %v\n", si.Directory)
	fmt.Printf("Project name: %v\n", si.ProjectName)

	err := os.Mkdir(".subsys", 0777)
	if err != nil {
		log.Fatal("This is already a subsys directory\n")
	}

	err = os.Mkdir(filepath.Join(".subsys", "snapshots"), 0777)
	if err != nil {
		return err
	}

	configFile, err := os.Create(filepath.Join(".subsys", "config.json"))
	if err != nil {
		return err
	}
	defer configFile.Close()

	config := utils.AssignmentConfig{
		ProjectName: si.ProjectName,
		Directory:   si.Directory,
	}
	encoder := json.NewEncoder(configFile)
	err = encoder.Encode(config)
	if err != nil {
		return err
	}

	ignoreFile, err := os.Create("subsysignore")
	if err != nil {
		return err
	}
	defer ignoreFile.Close()

	trackFile, err := os.Create(filepath.Join(".subsys", ".track"))
	if err != nil {
		return err
	}
	defer trackFile.Close()

	_, writeErr := ignoreFile.WriteString("# This file contains files to ignore in your snapshots\n.subsys\nsubsysignore\n.git")
	if writeErr != nil {
		return writeErr
	}

	return nil
}

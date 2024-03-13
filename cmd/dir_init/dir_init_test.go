package dirinit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDirectoryInitializer(t *testing.T) {
	initializer, err := NewDirectoryInitializer()

	if err != nil {
		t.Fatalf("Error creating DirectoryInitializer: %v", err)
	}

	if initializer.ProjectName == "" {
		t.Error("ProjectName is empty")
	}

	if initializer.Directory == "" {
		t.Error("Directory is empty")
	}
}

func TestInitialize(t *testing.T) {
	tempDir := t.TempDir()

	err := os.Chdir(tempDir)
	if err != nil {
		t.Errorf("Failed to change working directory: %v", err)
	}

	initializer := configInitializer{
		ProjectName: "test_project",
		Directory:   tempDir,
	}

	err = initializer.Initialize()

	if err != nil {
		t.Fatalf("Error initializing submission: %v", err)
	}

	configFilePath := filepath.Join(tempDir, ".subsys", "config.json")
	_, configErr := os.Stat(configFilePath)
	if configErr != nil {
		t.Fatalf("Error checking config file: %v", configErr)
	}

	ignoreFilePath := filepath.Join(tempDir, "subsysignore")
	_, ignoreErr := os.Stat(ignoreFilePath)
	if ignoreErr != nil {
		t.Fatalf("Error checking ignore file: %v", ignoreErr)
	}

	trackFilePath := filepath.Join(tempDir, ".subsys", ".track")
	_, trackErr := os.Stat(trackFilePath)
	if trackErr != nil {
		t.Fatalf("Error checking track file: %v", trackErr)
	}

	err = os.RemoveAll(".subsys")
	if err != nil {
		t.Fatal()
	}
}

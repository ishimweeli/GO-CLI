package dirconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	dirinit "amalitech.org/subsys/cmd/dir_init"
)

func setupConfigurator(t *testing.T) *Configurator {
	tempDir := t.TempDir()

	err := os.Chdir(tempDir)
	if err != nil {
		t.Errorf("Failed to change working directory: %v", err)
	}

	initializer, err := dirinit.NewDirectoryInitializer()
	if err != nil {
		t.Fatalf("NewDirectoryInitializer failed: %v", err)
	}

	err = initializer.Initialize()
	if err != nil {
		t.Fatalf("InitializeSubmission failed: %v", err)
	}

	return NewConfigurator()
}

func TestConfigureDirectory(t *testing.T) {
	configurator := setupConfigurator(t)

	os.Args = []string{"program", "config", "--code", "12345", "--student_id", "9876"}

	err := configurator.ConfigureDirectory()
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	err = os.RemoveAll(".subsys")
	if err != nil {
		t.Fatal(err)
	}
}

func TestWriteToConfigFile(t *testing.T) {
	configurator := setupConfigurator(t)

	configurator.AssCode = "54321"
	configurator.StudentID = "9876"

	err := configurator.writeToConfigFile()
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	configData, err := os.ReadFile(filepath.Join(".subsys", "config.json"))
	if err != nil {
		t.Fatalf("Error reading test config file: %v", err)
	}

	containsConfigData := `
"StudentID": "9876",
"AssignmentCode": "54321"
`
	if !strings.Contains(string(configData), containsConfigData) {
		t.Errorf("Expected config data:\n%s\nGot:\n%s", containsConfigData, string(configData))
	}

	err = os.RemoveAll(".subsys")
	if err != nil {
		t.Fatal(err)
	}
}

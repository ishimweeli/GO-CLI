package dirsubmission

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	dirconfig "amalitech.org/subsys/cmd/dir_config"
	dirinit "amalitech.org/subsys/cmd/dir_init"
	"amalitech.org/subsys/utils"
)

func SetupSubmissionTests(t *testing.T) SubmissionManager {
	tempDir := t.TempDir()

	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}

	initializer, err := dirinit.NewDirectoryInitializer()
	if err != nil {
		t.Fatalf("NewDirectoryInitializer failed: %v", err)
	}

	err = initializer.Initialize()
	if err != nil {
		t.Fatalf("InitializeDirectory failed: %v", err)
	}

	os.Args = []string{"program", "config", "--code", "12345", "--student_id", "9876"}

	configurator := dirconfig.NewConfigurator()

	err = configurator.ConfigureDirectory()
	if err != nil {
		t.Fatalf("Error configuring directory: %v\n", err)
	}

	config, err := utils.GetConfig()
	if err != nil {
		t.Fatalf("Couldn't get config file: %v \n", err)
	}

	return SubmissionManager{
		Config: config,
		Authorization: Auth{
			Password: "password",
		},
		SnapshotName: "test",
	}
}

func TestLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/admin/login" {
			t.Errorf("Expected path: /users/admin/login, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected method: POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"role": "student", "firstName": "John", "token": "testtoken"}`))
	}))
	defer server.Close()

	sm := SetupSubmissionTests(t)
	sm.ServerUrl = server.URL

	err := sm.Login()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if sm.Authorization.AccessToken != "testtoken" {
		t.Errorf("Expected access token: testtoken, got %s", sm.Authorization.AccessToken)
	}
}

func TestSubmitNoSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/submissions/student/create/submission/12345" {
			t.Errorf("Expected path: /submissions/student/create/submission/12345, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected method: POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Submission successful"}`))
	}))
	defer server.Close()

	sm := SetupSubmissionTests(t)

	sm.ServerUrl = server.URL

	err := sm.Submit()
	if !strings.Contains(err.Error(), "you don't have a snapshot named test") {
		t.Errorf("Expected error you don't have a snapshot named test but got: %v", err)
	}

	expectedResult := false
	if got := sm.success; got != expectedResult {
		t.Errorf("Expected success to be: %v, got %v", expectedResult, got)
	}

}

func TestSubmitSnapshotSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/submissions/student/create/submission/12345" {
			t.Errorf("Expected path: /submissions/student/create/submission/12345, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected method: POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Submission successful"}`))
	}))
	defer server.Close()

	sm := SetupSubmissionTests(t)

	sm.ServerUrl = server.URL

	f, err := os.Create(filepath.Join(".", ".subsys", "snapshots", "test.zip"))
	if err != nil {
		t.Errorf("Error creating test snapshot: %v", err)
	}
	defer f.Close()

	err = sm.Submit()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedResult := true
	if got := sm.success; got != expectedResult {
		t.Errorf("Expected success to be: %v, got %v", expectedResult, got)
	}
}

package dirclone

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func SetupCloneTests(t *testing.T) CloneManager {
	tempDir := t.TempDir()

	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}

	err = createTestZip()
	if err != nil {
		t.Errorf("Failed to create test zip file with error %v", err)
	}

	return CloneManager{
		LectureCode:  "LEC-1",
		SnapshotID:   "1",
		SubmissionID: "4",
		Authorization: Auth{
			Password: "password",
		},
	}
}

func TestNewCloneManager(t *testing.T) {
	manager := NewCloneManager()
	if manager == nil {
		t.Errorf("NewCloneManager returned nothing")
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

	sm := SetupCloneTests(t)
	sm.ServerUrl = server.URL

	err := sm.Login()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if sm.Authorization.AccessToken != "testtoken" {
		t.Errorf("Expected access token: testtoken, got %s", sm.Authorization.AccessToken)
	}
}

func TestLoginFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/admin/login" {
			t.Errorf("Expected path: /users/admin/login, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected method: POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"role": "student", "firstName": "John", "token": "testtoken"}`))
	}))
	defer server.Close()

	sm := SetupCloneTests(t)
	sm.ServerUrl = server.URL

	err := sm.Login()

	if err == nil {
		t.Errorf("Expected error: %v, but got none", err)
	}
}

func TestDownloadSnapshot(t *testing.T) {
	submissionID := "8"
	snapshotID := "12"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/submissions/lecturer/download/submissions" {
			t.Errorf("Expected path: /submissions/lecturer/download/submissions, but got %s", r.URL.Path)
		}

		if r.URL.Query().Get("submissionId") != submissionID {
			t.Errorf("Expected submissionId to be %v, but got %v", submissionID, r.URL.Query().Get("submissionId"))
		}

		if r.URL.Query().Get("snapshotId") != snapshotID {
			t.Errorf("Expected snapshotId to be %v, but got %v", snapshotID, r.URL.Query().Get("snapshotId"))
		}

		if r.Method != http.MethodGet {
			t.Errorf("Expected method: GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)

		f, err := os.Open("test.zip")
		if err != nil {
			t.Errorf("Error opening test snapshot: %v", err)
		}
		defer f.Close()

		w.Header().Set("Content-Type", "application/zip")

		_, err = io.Copy(w, f)
		if err != nil {
			t.Errorf("Error copying file contents to response: %v", err)
		}
	}))

	cm := SetupCloneTests(t)
	cm.ServerUrl = server.URL
	cm.SubmissionID = submissionID
	cm.SnapshotID = snapshotID

	err := cm.DownloadSnapshot()
	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}

}

func TestDownloadFail(t *testing.T) {
	submissionID := "8"
	snapshotID := "12"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)

		w.Header().Set("Content-Type", "application/zip")

	}))

	cm := SetupCloneTests(t)
	cm.ServerUrl = server.URL
	cm.SubmissionID = submissionID
	cm.SnapshotID = snapshotID

	err := cm.DownloadSnapshot()
	if err == nil {
		t.Fatalf("Expected an error %v, but got none", err)
	}

}

func createTestZip() error {
	zipFile, err := os.Create("test.zip")
	if err != nil {
		return fmt.Errorf("Error creating zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	file, err := zipWriter.Create("dummy.txt")
	if err != nil {
		return fmt.Errorf("Error creating dummy.txt in zip file: %v", err)
	}

	_, err = file.Write([]byte("This is a test file in the zip archive."))
	if err != nil {
		return fmt.Errorf("Error writing content to dummy.txt: %v", err)
	}

	return nil
}

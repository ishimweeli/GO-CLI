package dirsnap

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	dirinit "amalitech.org/subsys/cmd/dir_init"
	"amalitech.org/subsys/utils"
)

func SetupSnapshotManager(t *testing.T) *SnapshotManager {
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

	return NewSnapshotManager()
}

func TestCreateSnapshot(t *testing.T) {
	sm := SetupSnapshotManager(t)

	os.Args = []string{"program", "snap", "--name", "submission1"}

	err := sm.CreateSnapshot()
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(filepath.Join(".", ".subsys", "snapshots", sm.Name+".zip"))
	if err != nil {
		t.Error(err)
	}
}

func TestTrackChanges(t *testing.T) {
	sm := SetupSnapshotManager(t)

	changes, err := sm.TrackChanges("./")
	if err != nil {
		t.Error(err)
	}

	if len(changes) == 0 {
		t.Error("No changes were found")
	}

	for _, change := range changes {
		if change.Status != Added && change.Status != Modified && change.Status != Deleted {
			t.Error("Invalid change status:", change.Status)
		}
	}
}

func TestCompareSnapshot(t *testing.T) {
	sm := SetupSnapshotManager(t)

	os.Args = []string{"program", "snap", "--name", "submission1"}

	err := sm.CreateSnapshot()
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile("testfile.txt", []byte("Hello, World!"), 0644)
	if err != nil {
		t.Error(err)
	}

	changes, err := sm.TrackChanges(".")
	if err != nil {
		t.Error(err)
	}

	if len(changes) == 0 {
		t.Error("Expected to find changes but none found:", changes)
	}
}

func TestScanDir(t *testing.T) {
	sm := SetupSnapshotManager(t)

	os.Create("testFile.txt")

	changes, err := sm.scanDir(".")
	if err != nil {
		t.Error(err)
	}

	if len(changes) == 0 {
		t.Error("No changes were found")
	}

	for _, change := range changes {
		if change.Status != Added {
			t.Error("Invalid change status:", change.Status)
		}
	}
}

func TestCreateTracker(t *testing.T) {
	sm := SetupSnapshotManager(t)

	tracker, err := sm.createTracker("./")
	if err != nil {
		t.Error(err)
	}

	if len(tracker) != 0 {
		t.Error("Expected tracker file to be empty")
	}
}

func TestChecksum(t *testing.T) {
	sm := SetupSnapshotManager(t)

	file, err := os.Create("testfile.txt")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	checksum, err := sm.checksum("testfile.txt")
	if err != nil {
		t.Error(err)
	}

	if len(checksum) != 64 {
		t.Error("Invalid checksum:", checksum)
	}
}

func TestPrintChanges(t *testing.T) {
	sm := SetupSnapshotManager(t)

	changes := []FileChange{
		{Path: "./testfile1.txt", Status: Added},
		{Path: "./testfile2.txt", Status: Modified},
		{Path: "./testfile3.txt", Status: Deleted},
	}

	err := sm.printChanges(changes)
	if err != nil {
		t.Error(err)
	}
}

func TestCompress(t *testing.T) {
	sm := SetupSnapshotManager(t)

	os.Args = []string{"program", "snap", "--name", "submission1"}

	err := sm.CreateSnapshot()
	if err != nil {
		t.Error(err)
	}

	err = sm.compress()
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(filepath.Join(".", ".subsys", "snapshots", sm.Name+".zip"))
	if err != nil {
		t.Error(err)
	}

}

func TestGetSnapshotName(t *testing.T) {
	sm := SetupSnapshotManager(t)

	err := sm.GetSnapshotName()
	if err != nil {
		t.Error(err)
	}

	if !(len(sm.Name) > 0) {
		t.Error("A snapshot must have a name")
	}

	isNotSafe, _ := regexp.MatchString(`([&$\+,:;=\?@#\s<>\[\]\{\}[\/]|\\\^%])+`, sm.Name)
	if isNotSafe {
		t.Error("A snapshot's name must be a slug [https://medium.com/dailyjs/web-developer-playbook-slug-a6dcbe06c284](https://medium.com/dailyjs/web-developer-playbook-slug-a6dcbe06c284)")
	}
}

func TestGetIgnoreFile(t *testing.T) {
	SetupSnapshotManager(t)

	ignoredFiles, err := utils.GetIgnoredFiles(filepath.Join(".", "subsysignore"))

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(ignoredFiles) == 0 {
		t.Errorf("Expected data but ignore files, but found none")
	}
}

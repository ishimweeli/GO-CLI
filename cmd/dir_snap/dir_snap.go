package dirsnap

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"amalitech.org/subsys/utils"
)

type Status int

const (
	Added Status = iota
	Modified
	Deleted
)

type FileChange struct {
	Path   string
	Status Status
}

type SnapshotManager struct {
	Config utils.AssignmentConfig
	Name   string
}

func NewSnapshotManager() *SnapshotManager {
	config, err := utils.GetConfig()

	if err != nil {
		log.Fatalf("Couldn't get config file: %v \n", err)
		return nil
	}

	configData := utils.AssignmentConfig{
		ProjectName:    config.ProjectName,
		Directory:      config.Directory,
		AssignmentCode: config.AssignmentCode,
		StudentID:      config.StudentID,
	}

	return &SnapshotManager{
		Config: configData,
	}
}

func (sm *SnapshotManager) CreateSnapshot() error {
	err := sm.GetSnapshotName()
	if err != nil {
		log.Fatal(err)
	}

	changes, err := sm.TrackChanges("./")
	if err != nil {
		log.Fatal(err)
	}

	if len(changes) == 0 {
		return fmt.Errorf("no new changes")
	}

	sm.printChanges(changes)

	err = sm.compress()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Snapshot %s created successfully\n", sm.Name)

	return nil
}

func (sm *SnapshotManager) TrackChanges(dir string) ([]FileChange, error) {
	trackerPath := filepath.Join(dir, ".subsys", ".track")

	tracker, err := ioutil.ReadFile(trackerPath)

	if err == nil {
		return sm.compareSnapshot(tracker, dir)
	} else if os.IsNotExist(err) {
		changes, err := sm.scanDir(dir)
		if err != nil {
			return nil, err
		}

		tracker, err := sm.createTracker(dir)
		if err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(trackerPath, tracker, 0644)
		if err != nil {
			return nil, err
		}

		return changes, nil
	} else {
		return nil, err
	}
}

func (sm *SnapshotManager) compareSnapshot(snapshot []byte, dir string) ([]FileChange, error) {
	changes := make([]FileChange, 0)

	lines := strings.Split(string(snapshot), "\n")

	for _, line := range lines {
		file := strings.TrimSpace(line)

		currentPath := strings.Split(file, " ")

		_, err := os.Stat(currentPath[0])
		if err != nil {
			if os.IsNotExist(err) {
				changes = append(changes, FileChange{Path: file, Status: Deleted})
			} else {
				return nil, err
			}
		} else {

			currentHash, err := sm.checksum(currentPath[0])
			if err != nil {
				return nil, err
			}

			snapshotHash := ""
			for _, trackLine := range lines {
				trackerPath := strings.Split(trackLine, " ")
				if trackerPath[0] == currentPath[0] {
					snapshotHash = trackerPath[1]
				}
			}
			if currentHash != snapshotHash {
				changes = append(changes, FileChange{Path: file, Status: Modified})
			}
		}
	}

	newFiles, err := sm.scanDir(dir)
	if err != nil {
		return nil, err
	}

	trackComparison := fmt.Sprintf("%v", lines)

	for _, newFile := range newFiles {
		if !strings.Contains(trackComparison, newFile.Path) {
			changes = append(changes, newFile)
		}
	}

	if len(changes) > 0 {
		tracker, err := sm.createTracker(dir)
		if err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(filepath.Join(dir, ".subsys", ".track"), tracker, 0644)
		if err != nil {
			return nil, err
		}
	}

	return changes, nil
}

func (sm *SnapshotManager) scanDir(dir string) ([]FileChange, error) {
	changes := make([]FileChange, 0)

	ignoredFiles, err := utils.GetIgnoredFiles(filepath.Join(".", "subsysignore"))
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && !utils.Contains(ignoredFiles, path) {
			changes = append(changes, FileChange{Path: path, Status: Added})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return changes, nil
}

func (sm *SnapshotManager) createTracker(dir string) ([]byte, error) {
	hashes := make([]string, 0)

	ignoredFiles, err := utils.GetIgnoredFiles(filepath.Join(".", "subsysignore"))
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && !utils.Contains(ignoredFiles, path) {
			hash, err := sm.checksum(path)
			if err != nil {
				return err
			}

			hashes = append(hashes, fmt.Sprintf("%s %s", filepath.ToSlash(path), hash))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	snapshot := strings.Join(hashes, "\n")

	return []byte(snapshot), nil
}

func (sm *SnapshotManager) checksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (sm *SnapshotManager) printChanges(changes []FileChange) error {
	for _, change := range changes {
		if change.Path != "" {
			switch change.Status {
			case Added:
				fmt.Printf("Added: %s\n", change.Path)
			case Modified:
				fmt.Printf("Modified: %s\n", change.Path)
			case Deleted:
				fmt.Printf("Deleted: %s\n", change.Path)
			}
		}
	}
	return nil
}

func (sm *SnapshotManager) compress() error {
	f, err := os.Create(filepath.Join(".", ".subsys", "snapshots", sm.Name+".zip"))
	if err != nil {
		fmt.Printf("Error here\n")
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	ignoredFiles, err := utils.GetIgnoredFiles(filepath.Join(".", "subsysignore"))
	if err != nil {
		return err
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && !utils.Contains(ignoredFiles, path) {

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			header.Method = zip.Deflate

			header.Name, err = filepath.Rel(filepath.Dir("."), path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				header.Name += "/"
			}

			headerWriter, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(headerWriter, f)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (sm *SnapshotManager) GetSnapshotName() error {

	flags := flag.NewFlagSet("snap", flag.ContinueOnError)
	flags.StringVar(&sm.Name, "name", "", "Enter snapshot name")
	flags.Parse(os.Args[2:])

	if !(len(sm.Name) > 0) {
		err := errors.New("a snapshot must have a name")
		return err
	}

	isNotSafe, _ := regexp.MatchString(`([&$\+,:;=\?@#\s<>\[\]\{\}[\/]|\\\^%])+`, sm.Name)
	if isNotSafe {
		err := errors.New("a snapshot's name must be a slug, refer to this https://medium.com/dailyjs/web-developer-playbook-slug-a6dcbe06c284")
		return err
	}

	fmt.Printf("The snapshot name is %v\n", sm.Name)
	return nil
}

package dirclone

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"amalitech.org/subsys/utils"
)

type Auth struct {
	AccessToken string
	Password    string
}

type CloneManager struct {
	Authorization Auth
	ServerUrl     string
	success       bool
	LectureCode   string
	SubmissionID  string
	SnapshotID    string
}

func NewCloneManager() *CloneManager {
	return &CloneManager{
		ServerUrl: "https://gitinspired-rw-api.amalitech-dev.net/api",
	}
}

func (cm *CloneManager) CloneSnapshot() error {
	err := cm.getDataInteractively()
	if err != nil {
		return err
	}

	err = cm.Login()
	if err != nil {
		return err
	}

	err = cm.DownloadSnapshot()
	if err != nil {
		return err
	}

	return nil
}

func (cm *CloneManager) getDataInteractively() error {
	fmt.Print("Enter your lecture code: ")
	if err := utils.ReadInputUntilValid(&cm.LectureCode); err != nil {
		return err
	}

	fmt.Print("Enter the submission id: ")
	if err := utils.ReadInputUntilValid(&cm.SubmissionID); err != nil {
		return err
	}

	fmt.Print("Enter the snapshot id: ")
	if err := utils.ReadInputUntilValid(&cm.SnapshotID); err != nil {
		return err
	}

	fmt.Print("Enter your password: ")
	if err := utils.ReadInputUntilValid(&cm.Authorization.Password); err != nil {
		return err
	}

	if cm.SubmissionID == "" || cm.LectureCode == "" || cm.SnapshotID == "" {
		return errors.New("failed to get input data")
	}

	return nil
}

func (cm *CloneManager) Login() error {
	accessToken, err := utils.ServerLogin(cm.ServerUrl+"/users/admin/login", cm.LectureCode, cm.Authorization.Password)
	if err != nil {
		return err
	}

	cm.Authorization.AccessToken = accessToken
	return nil
}

func (cm *CloneManager) DownloadSnapshot() error {
	url := cm.ServerUrl + "/submissions/lecturer/download/submissions?submissionId=" + cm.SubmissionID + "&snapshotId=" + cm.SnapshotID
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return err
	}

	req.Header.Add("Authorization", "Bearer "+cm.Authorization.AccessToken)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		var submissionError utils.ServerError
		json.Unmarshal(body, &submissionError)
		err = errors.New(submissionError.Message)
		return err
	}

	dirName := "Submission-" + cm.SubmissionID + "-snap-" + cm.SnapshotID

	err = os.MkdirAll(dirName, 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.Chdir(dirName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, file := range reader.File {
		filePath := filepath.Join(".", file.Name)

		err = os.MkdirAll(filepath.Dir(filePath), 0777)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if file.FileInfo().IsDir() {
			continue
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer outFile.Close()

		rc, err := file.Open()
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer rc.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	cm.success = true
	fmt.Printf("Snapshot downloaded and extracted successfully to %v\n", dirName)
	return nil
}

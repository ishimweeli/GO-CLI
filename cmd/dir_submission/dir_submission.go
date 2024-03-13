package dirsubmission

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"amalitech.org/subsys/utils"
)

type Auth struct {
	AccessToken string
	Password    string
}

type SubmissionManager struct {
	Config        utils.AssignmentConfig
	Authorization Auth
	ServerUrl     string
	SnapshotName  string
	success       bool
}

type Login struct {
	Role      string `json:"role"`
	FirstName string `json:"firstName"`
	Token     string `json:"token"`
}

type serverError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type SubmissionResponse struct {
	Message string `json:"message"`
}

func NewSubmissionInitializer() *SubmissionManager {
	var snapshotName string
	var password string
	config, err := utils.GetConfig()
	fmt.Printf("Your Student ID: %v\n", config.StudentID)
	if err != nil {
		log.Fatalf("Couldn't get config file: %v \n", err)
		return nil
	}

	if config.StudentID == "" {
		log.Fatal("No Student ID found, first configure this directory")
		return nil
	}

	if config.AssignmentCode == "" {
		log.Fatal("No Assignment code found, first configure this directory")
		return nil
	}

	flags := flag.NewFlagSet("submit", flag.ContinueOnError)
	flags.StringVar(&snapshotName, "name", "", "Snapshot Name")
	flags.Parse(os.Args[2:])

	fmt.Printf("Enter your password: ")
	utils.ReadInputUntilValid(&password)

	return &SubmissionManager{
		Config: config,
		Authorization: Auth{
			Password: password,
		},
		ServerUrl:    "https://gitinspired-rw-api.amalitech-dev.net/api",
		SnapshotName: snapshotName,
	}
}

func (sm *SubmissionManager) SubmitSnapshots() error {
	err := sm.Login()
	if err != nil {
		return err
	}

	err = sm.Submit()
	if err != nil {
		return err
	}
	return nil
}

func (sm *SubmissionManager) Login() error {
	accessToken, err := utils.ServerLogin(sm.ServerUrl+"/users/admin/login", sm.Config.StudentID, sm.Authorization.Password)
	if err != nil {
		return err
	}

	sm.Authorization.AccessToken = accessToken
	return nil
}

func (sm *SubmissionManager) Submit() error {
	url := sm.ServerUrl + "/submissions/student/create/submission/" + sm.Config.AssignmentCode
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	fileFound := false
	validSnapshots := []string{}
	submittedSnaphots := []string{}
	err := filepath.Walk(filepath.Join(".", ".subsys", "snapshots"), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".zip" {

			fileName := strings.Split(info.Name(), ".")[0]
			validSnapshots = append(validSnapshots, fileName)

			if fileName == sm.SnapshotName || sm.SnapshotName == "" {
				fileFound = true
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				formFile, err := writer.CreateFormFile("snapshotArchive", filepath.Base(path))
				if err != nil {
					return err
				}
				_, err = io.Copy(formFile, file)
				if err != nil {
					return err
				}
				submittedSnaphots = append(submittedSnaphots, fileName)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return nil
	}

	if sm.SnapshotName != "" && !fileFound {
		err = fmt.Errorf("you don't have a snapshot named %s\nValid snapshots are: %s", sm.SnapshotName, strings.Join(validSnapshots, ", "))
		return err
	}

	if len(submittedSnaphots) == 0 {
		log.Fatal("You have no snapshots yet, first create a snapshot with the subsys snap command")
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+sm.Authorization.AccessToken)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		var submissionError serverError
		json.Unmarshal(body, &submissionError)
		err = errors.New(submissionError.Message)
		return err
	}

	sm.success = true

	fmt.Printf("Submitted snapshot(s): %v\n", strings.Join(submittedSnaphots, ","))

	var response SubmissionResponse
	json.Unmarshal(body, &response)

	fmt.Printf("%s\n", response.Message)

	return nil
}

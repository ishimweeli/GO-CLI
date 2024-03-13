package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Login struct {
	Role      string `json:"role"`
	FirstName string `json:"firstName"`
	Token     string `json:"token"`
}

func ServerLogin(url string, email string, password string) (token string, err error) {
	posturl := url

	requestBody := []byte(`{
		"email": "` + email + `",
  		"password": "` + password + `"
	}`)

	res, err := http.Post(posturl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		var authError ServerError
		json.Unmarshal(body, &authError)
		err = errors.New(authError.Message)
		return "", err
	}

	var data Login
	json.Unmarshal(body, &data)

	return data.Token, nil
}

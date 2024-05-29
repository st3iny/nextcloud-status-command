package ocs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const statusEndpoint string = "/ocs/v2.php/apps/user_status/api/v1/user_status/status"
const messageEndpoint string = "/ocs/v2.php/apps/user_status/api/v1/user_status/message?format=json"
const customMessageEndpoint string = "/ocs/v2.php/apps/user_status/api/v1/user_status/message/custom?format=json"

func getStatusEndpoint(user string) string {
	return fmt.Sprintf("/ocs/v2.php/apps/user_status/api/v1/statuses/%s", user)
}

type StatusMessage struct {
	ClearAt    int64  `json:"clearAt,omitempty"`
	Message    string `json:"message"`
	StatusIcon string `json:"statusIcon,omitempty"`
}

type Status struct {
	StatusType string `json:"statusType"`
}

type UserStatus struct {
	User    string
	Status  string
	Icon    string
	Message string
	ClearAt int64
}

func (a *Auth) Endpoint(url string) string {
	return fmt.Sprintf("%s%s", a.ServerBaseUrl, url)
}

func GetStatus(auth Auth) (*UserStatus, error) {
	req, err := http.NewRequest("GET", auth.Endpoint(getStatusEndpoint(auth.User)), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("OCS-APIRequest", "true")
	req.SetBasicAuth(auth.User, auth.Password)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var ocsResponse map[string]interface{}
	err = json.Unmarshal(resBody, &ocsResponse)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get status message: %s %s", res.Status, string(resBody))
	}

	data := ocsResponse["ocs"].(map[string]interface{})["data"].(map[string]interface{})
	status := UserStatus{
		User:   auth.User,
		Status: data["status"].(string),
	}

	switch data["message"].(type) {
	case string:
		status.Message = data["message"].(string)
	}

	switch data["icon"].(type) {
	case string:
		status.Icon = data["icon"].(string)
	}

	switch data["clearAt"].(type) {
	case float64:
		status.ClearAt = int64(data["clearAt"].(float64))
	}

	return &status, nil
}

func UpdateStatus(auth Auth, status Status) error {
	statusJson, err := json.Marshal(status)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(statusJson)
	req, err := http.NewRequest("PUT", auth.Endpoint(statusEndpoint), body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("OCS-APIRequest", "true")
	req.SetBasicAuth(auth.User, auth.Password)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	resBodyString := string(resBody)
	return fmt.Errorf("Failed to update status: %s %s", res.Status, resBodyString)
}

func UpdateStatusMessage(auth Auth, message StatusMessage) error {
	messageJson, err := json.Marshal(message)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(messageJson)
	req, err := http.NewRequest("PUT", auth.Endpoint(customMessageEndpoint), body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("OCS-APIRequest", "true")
	req.SetBasicAuth(auth.User, auth.Password)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	resBodyString := string(resBody)
	return fmt.Errorf("Failed to update status message: %s %s", res.Status, resBodyString)
}

func ClearStatusMessage(auth Auth) error {
	req, err := http.NewRequest("DELETE", auth.Endpoint(messageEndpoint), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("OCS-APIRequest", "true")
	req.SetBasicAuth(auth.User, auth.Password)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	resBodyString := string(resBody)
	return fmt.Errorf("Failed to clear status message: %s %s", res.Status, resBodyString)
}

package ocs

import (
	"encoding/json"
	"os"

	"github.com/adrg/xdg"
)

func LoadAuth() (Auth, error) {
	authPath, err := xdg.ConfigFile("nsc/auth.json")
	if err != nil {
		return Auth{}, err
	}

	authJson, err := os.ReadFile(authPath)
	if err != nil {
		return Auth{}, err
	}

	var auth Auth
	err = json.Unmarshal(authJson, &auth)
	if err != nil {
		return Auth{}, err
	}

	return auth, nil
}

func SaveAuth(auth Auth) error {
	authPath, err := xdg.ConfigFile("nsc/auth.json")
	if err != nil {
		return err
	}

	authJson, err := json.Marshal(&auth)
	if err != nil {
		return err
	}

	return os.WriteFile(authPath, authJson, 0600)
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

//Authenticate checks if email and password matches.
func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate the user against database

	user, err := app.Models.User.GetByEmail(requestPayload.Email)

	if err != nil {
		_ = app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)

	if err != nil || !valid {
		_ = app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// log authentication

	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))

	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "/t")
	request, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}
	return nil
}

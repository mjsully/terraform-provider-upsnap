package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mjsully/terraform-provider-upsnap/constants"
)

type APIClient struct {
	Client     *http.Client
	UpsnapHost string
	Token      string
	UserID     string
}

type AuthResponse struct {
	Token string `json:"token"` // adapt to actual API response
	User  struct {
		ID string `json:"id"`
	} `json:"record"`
}

func Authenticate(upsnapHost string, username string, password string) (AuthResponse, error) {

	apiUri := fmt.Sprintf("%s%s", upsnapHost, constants.AuthUri)

	bodyData := map[string]string{
		"identity": username,
		"password": password,
	}
	jsonBody, _ := json.Marshal(bodyData)

	// tflog.Debug(context.AfterFunc(), fmt.Sprintf("Attemping API request with the following data: %s, %s", apiUri, upsnapHost))

	resp, err := ApiInteraction(apiUri, "", "POST", bytes.NewBuffer(jsonBody))
	if err != nil {
		return AuthResponse{}, fmt.Errorf("failed to build auth request: %w", err)
	}
	defer resp.Body.Close()

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return AuthResponse{}, fmt.Errorf("failed to build auth request: %w", err)
	}

	return authResp, nil

}

func ApiInteraction(uri string, token string, method string, body io.Reader) (*http.Response, error) {

	fmt.Printf("Attemping API request with the following data: %s, %s", uri, token)

	request, _ := http.NewRequest(method, uri, body)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("API interaction failed: %w", err)
	}

	return response, nil

}

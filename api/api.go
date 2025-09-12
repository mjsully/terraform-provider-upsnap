package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mjsully/terraform-provider-upsnap/constants"
)

type AuthResponse struct {
	Token string `json:"token"` // adapt to actual API response
	User  struct {
		ID string `json:"id"`
	} `json:"record"`
}

func Authenticate(upsnapHost string, username string, password string) AuthResponse {

	apiUri := fmt.Sprintf("%s%s", upsnapHost, constants.AuthUri)

	bodyData := map[string]string{
		"identity": username,
		"password": password,
	}
	jsonBody, _ := json.Marshal(bodyData)

	resp, _ := ApiInteraction(apiUri, "", "POST", bytes.NewBuffer(jsonBody))
	defer resp.Body.Close()

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		panic(err)
	}

	return authResp

}

func ApiInteraction(uri string, token string, method string, body io.Reader) (*http.Response, error) {

	req, _ := http.NewRequest(method, uri, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	return resp, err

}

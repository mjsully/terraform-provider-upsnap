package upsnap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

type APIResponse struct {
	ID string `json:"id"` // adapt to actual API response
}

type DeviceResponse struct {
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Mac     string `json:"mac"`
	Netmask string `json:"netmask"`
}

// type PermissionList struct {
// 	Items []struct {
// 		Delete []string `json:"delete"`
// 		Power  []string `json:"power"`
// 		Read   []string `json:"read"`
// 		Update []string `json:"update"`
// 	} `json:"items"`
// }

func Authenticate(upsnapHost string, username string, password string) AuthResponse {

	apiUrl := fmt.Sprintf("%s%s", upsnapHost, constants.AuthUri)

	bodyData := map[string]string{
		"identity": username,
		"password": password,
	}
	jsonBody, _ := json.Marshal(bodyData)

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("auth failed: %s", resp.Status))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		panic(err)
	}

	return authResp

}

// func GetPermissions(upsnapHost string, apiToken string) PermissionList {

// 	apiUrl := fmt.Sprintf("%s%s", upsnapHost, constants.PermissionsUri)

// 	req, _ := http.NewRequest("GET", apiUrl, nil)
// 	req.Header.Set("Authorization", "Bearer "+apiToken)
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		panic(fmt.Errorf("auth failed: %s", resp.Status))
// 	}

// 	var apiResp PermissionList
// 	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
// 		panic(err)
// 	}

// 	// deleteList := apiResp.Items[0].Delete
// 	// deleteList = append(deleteList, "my-new-id")

// 	// fmt.Println(deleteList)

// 	return apiResp

// }

func Provider() *schema.Provider {
	return &schema.Provider{

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"upsnap_host": {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"upsnap_device": resourceDevice(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

// ConfigureContextFunc runs when provider is initialized
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	upsnapHost := d.Get("upsnap_host").(string)

	token := Authenticate(upsnapHost, username, password)

	apiClient := &APIClient{
		Client:     &http.Client{},
		UpsnapHost: upsnapHost,
		Token:      token.Token,
		UserID:     token.User.ID,
	}

	return apiClient, nil
}

func resourceDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			insertUri := fmt.Sprintf("%s%s", apiData.UpsnapHost, constants.InsertUri)

			bodyData := map[string]string{
				"name":    d.Get("name").(string),
				"ip":      d.Get("ip").(string),
				"mac":     d.Get("mac").(string),
				"netmask": d.Get("netmask").(string),
			}
			jsonBody, _ := json.Marshal(bodyData)

			req, _ := http.NewRequest("POST", insertUri, bytes.NewBuffer(jsonBody))
			req.Header.Set("Authorization", "Bearer "+apiData.Token)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				panic(fmt.Errorf("auth failed: %s", resp.Status))
			}

			var apiResp APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				panic(err)
			}

			// permissions := GetPermissions(apiData.UpsnapHost, apiData.Token)

			// deleteList := permissions.Items[0].Delete
			// updateList := permissions.Items[0].Update
			// readList := permissions.Items[0].Read
			// powerList := permissions.Items[0].Power
			// deleteList = append(deleteList, apiResp.ID)
			// updateList = append(updateList, apiResp.ID)
			// readList = append(readList, apiResp.ID)
			// powerList = append(powerList, apiResp.ID)

			// permissionsUri := fmt.Sprintf("%s%s", apiData.UpsnapHost, constants.PermissionsUri)

			// permissionsBodyData := map[string]interface{}{
			// 	"delete": deleteList,
			// 	"update": updateList,
			// 	"read":   readList,
			// 	"power":  powerList,
			// }
			// permissionsJsonBody, _ := json.Marshal(permissionsBodyData)

			// req, _ = http.NewRequest("POST", permissionsUri, bytes.NewBuffer(permissionsJsonBody))

			// req.Header.Set("Authorization", "Bearer "+apiData.Token)
			// req.Header.Set("Content-Type", "application/json")

			// client = &http.Client{}
			// resp, err = client.Do(req)
			// if err != nil {
			// 	panic(err)
			// }
			// defer resp.Body.Close()

			// if resp.StatusCode != http.StatusOK {
			// 	panic(fmt.Errorf("auth failed: %s", resp.Status))
			// }

			d.SetId(apiResp.ID)
			return nil
		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			id := d.Id()

			apiUri := fmt.Sprintf("%s%s/%s", apiData.UpsnapHost, constants.InsertUri, id)

			req, _ := http.NewRequest("GET", apiUri, nil)
			req.Header.Set("Authorization", "Bearer "+apiData.Token)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				panic(fmt.Errorf("auth failed: %s", resp.Status))
			}

			var apiResp DeviceResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				panic(err)
			}

			d.Set("name", apiResp.Name)
			d.Set("ip", apiResp.IP)
			d.Set("mac", apiResp.Mac)
			d.Set("netmask", apiResp.Netmask)

			return nil
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			id := d.Id()

			insertUri := fmt.Sprintf("%s%s/%s", apiData.UpsnapHost, constants.InsertUri, id)

			bodyData := map[string]string{
				"name":    d.Get("name").(string),
				"ip":      d.Get("ip").(string),
				"mac":     d.Get("mac").(string),
				"netmask": d.Get("netmask").(string),
			}
			jsonBody, _ := json.Marshal(bodyData)

			req, _ := http.NewRequest("PATCH", insertUri, bytes.NewBuffer(jsonBody))
			req.Header.Set("Authorization", "Bearer "+apiData.Token)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				panic(fmt.Errorf("auth failed: %s", resp.Status))
			}

			var apiResp APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				panic(err)
			}
			return nil
		},
		DeleteContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			id := d.Id()

			insertUri := fmt.Sprintf("%s%s/%s", apiData.UpsnapHost, constants.InsertUri, id)

			req, _ := http.NewRequest("DELETE", insertUri, nil)
			req.Header.Set("Authorization", "Bearer "+apiData.Token)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			d.SetId("")
			return nil
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mac": {
				Type:     schema.TypeString,
				Required: true,
			},
			"netmask": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

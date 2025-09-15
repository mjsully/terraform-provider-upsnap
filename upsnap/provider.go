package upsnap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mjsully/terraform-provider-upsnap/api"
	"github.com/mjsully/terraform-provider-upsnap/constants"
)

type APIClient struct {
	Client     *http.Client
	UpsnapHost string
	Token      string
	UserID     string
}

type APIResponse struct {
	ID string `json:"id"`
}

type DeviceResponse struct {
	Name        string   `json:"name"`
	IP          string   `json:"ip"`
	Mac         string   `json:"mac"`
	Netmask     string   `json:"netmask"`
	Description string   `json:"description"`
	Link        string   `json:"link"`
	Groups      []string `json:"groups"`
}

type DeviceGroupResponse struct {
	Name string `json:"name"`
}

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
			"upsnap_device":       resourceDevice(),
			"upsnap_device_group": resourceDeviceGroup(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

// ConfigureContextFunc runs when provider is initialized
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	upsnapHost := d.Get("upsnap_host").(string)

	token := api.Authenticate(upsnapHost, username, password)

	apiClient := &APIClient{
		Client:     &http.Client{},
		UpsnapHost: upsnapHost,
		Token:      token.Token,
		UserID:     token.User.ID,
	}

	return apiClient, nil
}

func expandStringList(list []interface{}) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}

func resourceDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			apiUri := fmt.Sprintf("%s%s", apiData.UpsnapHost, constants.DeviceUri)

			bodyData := map[string]interface{}{
				"name":    d.Get("name").(string),
				"ip":      d.Get("ip").(string),
				"mac":     d.Get("mac").(string),
				"netmask": d.Get("netmask").(string),
			}
			if v, ok := d.GetOk("description"); ok {
				bodyData["description"] = v.(string)
			}
			if v, ok := d.GetOk("link"); ok {
				bodyData["link"] = v.(string)
				bodyData["link_open"] = "new_tab"
			}
			if v, ok := d.GetOk("groups"); ok {
				rawGroups := v.([]interface{})
				stringGroups := make([]string, len(rawGroups))
				for i, value := range rawGroups {
					stringGroups[i] = value.(string)
				}
				bodyData["groups"] = stringGroups
			}
			jsonBody, _ := json.Marshal(bodyData)

			resp, _ := api.ApiInteraction(apiUri, apiData.Token, "POST", bytes.NewBuffer(jsonBody))
			defer resp.Body.Close()

			var apiResp APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				panic(err)
			}

			d.SetId(apiResp.ID)
			return nil
		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			id := d.Id()
			apiUri := fmt.Sprintf("%s%s/%s", apiData.UpsnapHost, constants.DeviceUri, id)

			resp, _ := api.ApiInteraction(apiUri, apiData.Token, "GET", nil)
			defer resp.Body.Close()

			var apiResp DeviceResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				panic(err)
			}

			d.Set("name", apiResp.Name)
			d.Set("ip", apiResp.IP)
			d.Set("mac", apiResp.Mac)
			d.Set("netmask", apiResp.Netmask)
			d.Set("description", apiResp.Description)
			d.Set("link", apiResp.Link)
			if apiResp.Groups != nil {
				groupsList := make([]interface{}, len(apiResp.Groups))
				for i, group := range apiResp.Groups {
					groupsList[i] = group
				}
				d.Set("groups", groupsList)
			}

			return nil
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			id := d.Id()
			apiUri := fmt.Sprintf("%s%s/%s", apiData.UpsnapHost, constants.DeviceUri, id)

			bodyData := map[string]interface{}{
				"name":    d.Get("name").(string),
				"ip":      d.Get("ip").(string),
				"mac":     d.Get("mac").(string),
				"netmask": d.Get("netmask").(string),
			}
			if v, ok := d.GetOk("description"); ok {
				bodyData["description"] = v.(string)
			} else {
				bodyData["description"] = ""
			}
			if v, ok := d.GetOk("link"); ok {
				bodyData["link"] = v.(string)
				bodyData["link_open"] = "new_tab"
			} else {
				bodyData["link"] = ""
			}
			if v, ok := d.GetOk("groups"); ok {
				rawGroups := v.([]interface{})
				stringGroups := make([]string, len(rawGroups))
				for i, value := range rawGroups {
					stringGroups[i] = value.(string)
				}
				bodyData["groups"] = stringGroups
			}
			jsonBody, _ := json.Marshal(bodyData)

			resp, _ := api.ApiInteraction(apiUri, apiData.Token, "PATCH", bytes.NewBuffer(jsonBody))
			defer resp.Body.Close()

			var apiResp APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				panic(err)
			}
			return nil
		},
		DeleteContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			id := d.Id()

			apiUri := fmt.Sprintf("%s%s/%s", apiData.UpsnapHost, constants.DeviceUri, id)

			resp, _ := api.ApiInteraction(apiUri, apiData.Token, "DELETE", nil)
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"link": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDeviceGroup() *schema.Resource {

	return &schema.Resource{

		CreateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)
			apiUri := fmt.Sprintf("%s%s", apiData.UpsnapHost, constants.DeviceGroupUri)

			bodyData := map[string]string{
				"name": d.Get("name").(string),
			}
			jsonBody, _ := json.Marshal(bodyData)

			resp, _ := api.ApiInteraction(apiUri, apiData.Token, "POST", bytes.NewBuffer(jsonBody))
			defer resp.Body.Close()

			var apiResp APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				panic(err)
			}

			d.SetId(apiResp.ID)

			return nil

		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			id := d.Id()
			apiUri := fmt.Sprintf("%s%s/%s", apiData.UpsnapHost, constants.DeviceGroupUri, id)

			resp, _ := api.ApiInteraction(apiUri, apiData.Token, "GET", nil)
			defer resp.Body.Close()

			var apiResp DeviceGroupResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				panic(err)
			}

			d.Set("name", apiResp.Name)

			return nil

		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			id := d.Id()

			apiUri := fmt.Sprintf("%s%s/%s", apiData.UpsnapHost, constants.DeviceGroupUri, id)

			bodyData := map[string]string{
				"name": d.Get("name").(string),
			}
			jsonBody, _ := json.Marshal(bodyData)

			resp, _ := api.ApiInteraction(apiUri, apiData.Token, "PATCH", bytes.NewBuffer(jsonBody))
			defer resp.Body.Close()

			var apiResp APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				panic(err)
			}
			return nil

		},
		DeleteContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

			apiData := m.(*APIClient)

			id := d.Id()

			apiUri := fmt.Sprintf("%s%s/%s", apiData.UpsnapHost, constants.DeviceGroupUri, id)

			resp, _ := api.ApiInteraction(apiUri, apiData.Token, "DELETE", nil)
			defer resp.Body.Close()

			d.SetId("")

			return nil

		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

}

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mjsully/terraform-provider-upsnap/api"
	"github.com/mjsully/terraform-provider-upsnap/constants"
)

var (
	_ resource.Resource              = &deviceResource{}
	_ resource.ResourceWithConfigure = &deviceResource{}
)

type DeviceResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	IP          string   `json:"ip"`
	Mac         string   `json:"mac"`
	Netmask     string   `json:"netmask"`
	Description string   `json:"description"`
	Link        string   `json:"link"`
	Groups      []string `json:"groups"`
}

func NewDeviceResource() resource.Resource {
	return &deviceResource{}
}

type deviceResource struct {
	client *api.APIClient
}

type deviceResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	IP          types.String   `tfsdk:"ip"`
	Mac         types.String   `tfsdk:"mac"`
	Netmask     types.String   `tfsdk:"netmask"`
	Description types.String   `tfsdk:"description"`
	Link        types.String   `tfsdk:"link"`
	Groups      []types.String `tfsdk:"groups"`
	// updated types.String `tfsdk:"updated"`
}

// Metadata returns the resource type name.
func (r *deviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *deviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"ip": schema.StringAttribute{
				Required: true,
			},
			"mac": schema.StringAttribute{
				Required: true,
			},
			"netmask": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"link": schema.StringAttribute{
				Optional: true,
			},
			"groups": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *deviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *upsnap.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *deviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan deviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiUri := fmt.Sprintf("%s%s", r.client.UpsnapHost, constants.DeviceUri)

	var groupsList []string
	for _, group := range plan.Groups {
		groupsList = append(groupsList, group.ValueString())
	}

	bodyData := map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"ip":          plan.IP.ValueString(),
		"mac":         plan.Mac.ValueString(),
		"netmask":     plan.Netmask.ValueString(),
		"description": plan.Description.ValueString(),
		"link":        plan.Link.ValueString(),
		"groups":      groupsList,
	}
	jsonBody, _ := json.Marshal(bodyData)

	response, err := api.ApiInteraction(apiUri, r.client.Token, "POST", bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating resource",
			err.Error(),
		)
	}
	defer response.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Failed to decode response",
			err.Error(),
		)
	}

	plan.ID = types.StringValue(apiResp.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state deviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	apiUri := fmt.Sprintf("%s%s/%s", r.client.UpsnapHost, constants.DeviceUri, id)

	response, err := api.ApiInteraction(apiUri, r.client.Token, "GET", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving device",
			err.Error(),
		)
	}
	defer response.Body.Close()

	var apiResp DeviceResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Error decoding device response",
			err.Error(),
		)
	}

	groups := make([]types.String, len(apiResp.Groups))
	for i, g := range apiResp.Groups {
		groups[i] = types.StringValue(g)
	}

	state.ID = types.StringValue(id)
	state.Name = types.StringValue(apiResp.Name)
	state.IP = types.StringValue(apiResp.IP)
	state.Mac = types.StringValue(apiResp.Mac)
	state.Netmask = types.StringValue(apiResp.Netmask)
	state.Description = types.StringValue(apiResp.Description)
	state.Link = types.StringValue(apiResp.Link)
	state.Groups = groups

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var priorState deviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan deviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := priorState.ID.ValueString()

	apiUri := fmt.Sprintf("%s%s/%s", r.client.UpsnapHost, constants.DeviceUri, id)

	var groupsList []string
	for _, group := range plan.Groups {
		groupsList = append(groupsList, group.ValueString())
	}

	bodyData := map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"ip":          plan.IP.ValueString(),
		"mac":         plan.Mac.ValueString(),
		"netmask":     plan.Netmask.ValueString(),
		"description": plan.Description.ValueString(),
		"link":        plan.Link.ValueString(),
		"groups":      groupsList,
	}
	jsonBody, _ := json.Marshal(bodyData)

	response, err := api.ApiInteraction(apiUri, r.client.Token, "PATCH", bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating device",
			err.Error(),
		)
	}
	defer response.Body.Close()

	var apiResp DeviceResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Error decoding device response",
			err.Error(),
		)
	}

	ctx = tflog.SetField(ctx, "dbid", apiResp.ID)
	ctx = tflog.SetField(ctx, "name", apiResp.Name)
	ctx = tflog.SetField(ctx, "ip", apiResp.IP)
	ctx = tflog.SetField(ctx, "mac", apiResp.Mac)
	ctx = tflog.SetField(ctx, "netmask", apiResp.Netmask)
	ctx = tflog.SetField(ctx, "description", apiResp.Description)
	ctx = tflog.SetField(ctx, "link", apiResp.Link)

	tflog.Debug(ctx, "Attemped Device update")

	plan.ID = types.StringValue(id)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state deviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	apiUri := fmt.Sprintf("%s%s/%s", r.client.UpsnapHost, constants.DeviceUri, id)

	response, err := api.ApiInteraction(apiUri, r.client.Token, "DELETE", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting device",
			err.Error(),
		)
		return
	}
	defer response.Body.Close()

}

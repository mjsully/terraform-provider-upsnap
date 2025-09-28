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
	_ resource.Resource              = &deviceGroupResource{}
	_ resource.ResourceWithConfigure = &deviceGroupResource{}
)

type APIResponse struct {
	ID string `json:"id"`
}

type DeviceGroupResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewDeviceGroupResource() resource.Resource {
	return &deviceGroupResource{}
}

type deviceGroupResource struct {
	client *api.APIClient
}

type deviceGroupResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	// updated types.String `tfsdk:"updated"`
}

// Metadata returns the resource type name.
func (r *deviceGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_group"
}

func (r *deviceGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			// "updated": schema.StringAttribute{
			// 	Computed: true,
			// },
			"name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *deviceGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deviceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan deviceGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiUri := fmt.Sprintf("%s%s", r.client.UpsnapHost, constants.DeviceGroupUri)

	bodyData := map[string]string{
		"name": plan.Name.ValueString(),
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

func (r *deviceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state deviceGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	apiUri := fmt.Sprintf("%s%s/%s", r.client.UpsnapHost, constants.DeviceGroupUri, id)

	response, err := api.ApiInteraction(apiUri, r.client.Token, "GET", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving device group",
			err.Error(),
		)
	}
	defer response.Body.Close()

	var apiResp DeviceGroupResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Error decoding device group response",
			err.Error(),
		)
	}

	state.ID = types.StringValue(id)
	state.Name = types.StringValue(apiResp.Name)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *deviceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var priorState deviceGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan deviceGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := priorState.ID.ValueString()

	apiUri := fmt.Sprintf("%s%s/%s", r.client.UpsnapHost, constants.DeviceGroupUri, id)

	bodyData := map[string]string{
		"name": plan.Name.ValueString(),
	}
	jsonBody, _ := json.Marshal(bodyData)

	response, err := api.ApiInteraction(apiUri, r.client.Token, "PATCH", bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating device group",
			err.Error(),
		)
	}
	defer response.Body.Close()

	var apiResp DeviceGroupResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError(
			"Error decoding device group response",
			err.Error(),
		)
	}

	ctx = tflog.SetField(ctx, "dbid", apiResp.ID)
	ctx = tflog.SetField(ctx, "name", apiResp.Name)

	tflog.Debug(ctx, "Attemped DeviceGroup update")

	plan.ID = types.StringValue(id)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *deviceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state deviceGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	apiUri := fmt.Sprintf("%s%s/%s", r.client.UpsnapHost, constants.DeviceGroupUri, id)

	response, err := api.ApiInteraction(apiUri, r.client.Token, "DELETE", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting device group",
			err.Error(),
		)
		return
	}
	defer response.Body.Close()

}

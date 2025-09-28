package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mjsully/terraform-provider-upsnap/api"
)

// ScaffoldingProvider defines the provider implementation.
type UpsnapProvider struct {
	version string
}

type UpsnapProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Host     types.String `tfsdk:"host"`
}

var _ provider.Provider = &UpsnapProvider{}

func (p *UpsnapProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "UpSnap username",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "UpSnap password",
				Required:            true,
				Sensitive:           true,
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "UpSnap hostname",
				Required:            true,
			},
		},
	}
}

func (p *UpsnapProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data UpsnapProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	username := data.Username.ValueString()
	password := data.Password.ValueString()
	upsnapHost := data.Host.ValueString()

	ctx = tflog.SetField(ctx, "Host", upsnapHost)
	ctx = tflog.SetField(ctx, "Username", username)

	tflog.Debug(ctx, "Creating UpSnap API client")

	token, err := api.Authenticate(upsnapHost, username, password)
	if err != nil {
		resp.Diagnostics.AddError("Unable to authenticate", err.Error())
	}

	apiClient := &api.APIClient{
		Client:     &http.Client{},
		UpsnapHost: upsnapHost,
		Token:      token.Token,
		UserID:     token.User.ID,
	}

	resp.ResourceData = apiClient
}

func (p *UpsnapProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *UpsnapProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "upsnap"
	resp.Version = p.version
}

func (p *UpsnapProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDeviceGroupResource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UpsnapProvider{
			version: version,
		}
	}
}

package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure [ProjectDataSource] fully satisfies framework interfaces.
var (
	_ datasource.DataSource              = &ProjectDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectDataSource{}
)

func NewCoffeesDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource defines the [manifest.KindProject] data source implementation.
type ProjectDataSource struct {
	client *sdkClient
}

// Metadata returns the data source type name.
func (d *ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema implements [datasource.DataSource.Schema] function.
func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	description := "[Project configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#project)"
	resp.Schema = schema.Schema{
		Description:         description,
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"name": metadataNameAttr(),
			"display_name": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, diags := d.client.GetProject(ctx, config.Name)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := newProjectDataSourceModelFromManifest(project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *ProjectDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdkClient)
	if !ok {
		addInvalidSDKClientTypeDiag(&resp.Diagnostics, req.ProviderData)
		return
	}
	d.client = client
}

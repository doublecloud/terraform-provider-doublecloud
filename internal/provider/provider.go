package provider

import (
	"context"
	"errors"
	"os"

	dc "github.com/doublecloud/go-sdk"
	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &DoubleCloudProvider{}

// DoubleCloudProvider defines the provider implementation.
type DoubleCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	overrideEndpoint string
}

// DoubleCloudProviderModel describes the provider data model.
type DoubleCloudProviderModel struct {
	AuthorizedKey types.String `tfsdk:"authorized_key"`
	Endpoint      types.String `tfsdk:"endpoint"`
	TokenURL      types.String `tfsdk:"token_url"`
}

type Config struct {
	Credentials *dc.Credentials
	ProjectId   string
	Endpoint    string

	overrideEndpoint string

	ctx context.Context

	sdk *dc.SDK
}

func (c *Config) init(ctx context.Context) error {
	cfg := dc.Config{
		Credentials: *c.Credentials,
		Endpoint:    c.Endpoint,
	}

	if c.overrideEndpoint != "" {
		cfg.Endpoint = c.overrideEndpoint
		cfg.OverrideEndpoint = true
		cfg.Plaintext = true
	}

	sdk, err := dc.Build(ctx, cfg)
	if err != nil {
		return err
	}
	c.sdk = sdk
	c.ctx = ctx
	return nil
}

func (p *DoubleCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "doublecloud"
	resp.Version = p.version
}

func (p *DoubleCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"authorized_key": schema.StringAttribute{
				MarkdownDescription: "Path to authorized key",
				Optional:            true,
				Sensitive:           true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "API endpoint",
				Optional:            true,
			},
			"token_url": schema.StringAttribute{
				MarkdownDescription: "URL for token resolver",
				Optional:            true,
			},
		},
	}
}
func configureCredentials(data *DoubleCloudProviderModel) (dc.Credentials, error) {
	envToken := os.Getenv("DC_TOKEN")
	if envToken != "" {
		return dc.NewIAMTokenCredentials(envToken), nil
	}

	var key *iamkey.Key
	var err error

	envKey := os.Getenv("DC_AUTHKEY")

	if data.AuthorizedKey.IsNull() && envKey == "" {
		return nil, errors.New("please specify one of auth methods for Double.Cloud")
	}
	if envKey != "" {
		key, err = iamkey.ReadFromJSONFile(envKey)
	} else {
		key, err = iamkey.ReadFromJSONBytes([]byte(data.AuthorizedKey.ValueString()))
	}
	if err != nil {
		return nil, err
	}
	if data.Endpoint.ValueString() != "" {
		os.Setenv("DOUBLE_CLOUD_TOKEN_URL", data.TokenURL.ValueString())
	}
	return dc.ServiceAccountKey(key)
}

func (p *DoubleCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data DoubleCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	creds, err := configureCredentials(&data)
	if err != nil {
		resp.Diagnostics.AddError("failed to use credentials", err.Error())
		return
	}
	conf := &Config{
		Credentials: &creds,
		Endpoint:    data.Endpoint.ValueString(),

		overrideEndpoint: p.overrideEndpoint,
	}
	err = conf.init(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to init client", err.Error())
	}

	// TODO: forward conf struct instead of sdk
	resp.DataSourceData = conf.sdk
	resp.ResourceData = conf.sdk
}

func (p *DoubleCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNetworkResource,
		NewWorkbookResource,
		NewKafkaClusterResource,
		NewTransferResource,
		NewTransferEndpointResource,
		NewClickhouseClusterResource,
		NewNetworkConnectionResource,
		NewNetworkConnectionAccepterResource,
	}
}

func (p *DoubleCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNetworkDataSource,
		// NewWorkbookDataSource,
		NewKafkaDataSource,
		NewTransferDataSource,
		// NewTransferEndpointDataSource,
		NewClickhouseDataSource,
		NewNetworkConnectionDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DoubleCloudProvider{
			version: version,
		}
	}
}

package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointMssqlSourceSettings struct {
	Host              types.String            `tfsdk:"host"`
	Port              types.Int64             `tfsdk:"port"`
	Database          types.String            `tfsdk:"database"`
	Username          types.String            `tfsdk:"username"`
	Password          types.String            `tfsdk:"password"`
	ReplicationMethod types.String            `tfsdk:"replication_method"`
	SSLMethod         *endpointMssqlSSLMethod `tfsdk:"ssl_method"`
}

type endpointMssqlSSLMethod struct {
	Unencrypted         *endpointMssqlSSLMethodUnencrypted         `tfsdk:"unencrypted"`
	EncryptedTrusted    *endpointMssqlSSLMethodEncryptedTrusted    `tfsdk:"encrypted_trusted"`
	EncryptedVerifyCert *endpointMssqlSSLMethodEncryptedVerifyCert `tfsdk:"encrypted_verify_cert"`
}

type endpointMssqlSSLMethodUnencrypted struct{}

type endpointMssqlSSLMethodEncryptedTrusted struct{}

type endpointMssqlSSLMethodEncryptedVerifyCert struct {
	HostNameInCertificate types.String `tfsdk:"host_name_in_certificate"`
}

func transferEndpointMssqlSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The hostname of the database.",
			},
			"port": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The port of the database.",
			},
			"database": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name of the database.",
			},
			"username": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The username which is used to access the database.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The password associated with the username.",
			},
			"replication_method": schema.StringAttribute{
				Optional:            true,
				Validators:          []validator.String{transferEndpointMssqlReplicationMethodValidator()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The replication method used for extracting data from the database. STANDARD replication requires no setup on the DB side but will not be able to represent deletions incrementally. CDC uses {TBC} to detect inserts, updates, and deletes. This needs to be configured on the source database itself.",
			},
		},
		Blocks: map[string]schema.Block{
			"ssl_method": transferEndpointMssqlSSLMethodSchema(),
		},
	}
}

func transferEndpointMssqlSSLMethodSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"unencrypted": schema.SingleNestedBlock{
				MarkdownDescription: "Data transfer will not be encrypted.",
			},
			"encrypted_trusted": schema.SingleNestedBlock{
				MarkdownDescription: "Use the certificate provided by the server without verification. (For testing purposes only!)",
			},
			"encrypted_verify_cert": transferEndpointMssqlSSLMethodEncryptedVerifyCertSchema(),
		},
		MarkdownDescription: "",
	}
}

func transferEndpointMssqlSSLMethodEncryptedVerifyCertSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"host_name_in_certificate": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Specifies the host name of the server. The value of this property must match the subject property of the certificate.",
			},
		},
		MarkdownDescription: "Verify and use the certificate provided by the server.",
	}
}

func transferEndpointMssqlReplicationMethodValidator() validator.String {
	names := make([]string, len(endpoint_airbyte.MSSQLSource_MSSQLReplicationMethod_name))
	for i, name := range endpoint_airbyte.MSSQLSource_MSSQLReplicationMethod_name {
		names[i] = name
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func (m *endpointMssqlSourceSettings) convert() (*transfer.EndpointSettings_MssqlSource, diag.Diagnostics) {
	var diags diag.Diagnostics
	res := &endpoint_airbyte.MSSQLSource{}

	res.Host = m.Host.ValueString()
	res.Port = m.Port.ValueInt64()
	res.Database = m.Database.ValueString()
	res.Username = m.Username.ValueString()
	res.Password = m.Password.ValueString()
	res.ReplicationMethod = endpoint_airbyte.MSSQLSource_MSSQLReplicationMethod(
		endpoint_airbyte.MSSQLSource_MSSQLReplicationMethod_value[m.ReplicationMethod.ValueString()],
	)

	if m.SSLMethod != nil {
		sslMethod, d := m.SSLMethod.convert()
		diags.Append(d...)
		res.SslMethod = sslMethod
	}

	return &transfer.EndpointSettings_MssqlSource{
		MssqlSource: res,
	}, diags
}

func (m *endpointMssqlSSLMethod) convert() (*endpoint_airbyte.MSSQLSource_SSLConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	res := &endpoint_airbyte.MSSQLSource_SSLConfig{}
	switch {
	case m.Unencrypted != nil:
		res.SslMethod = &endpoint_airbyte.MSSQLSource_SSLConfig_Unencrypted{
			Unencrypted: new(endpoint_airbyte.MSSQLSource_SSLUnencrypted),
		}
	case m.EncryptedTrusted != nil:
		res.SslMethod = &endpoint_airbyte.MSSQLSource_SSLConfig_EncryptedTrustServerCertificate{
			EncryptedTrustServerCertificate: new(endpoint_airbyte.MSSQLSource_SSLEncryptedTrusted),
		}
	case m.EncryptedVerifyCert != nil:
		encryptedVerifyCert, d := m.EncryptedVerifyCert.convert()
		diags.Append(d...)
		res.SslMethod = encryptedVerifyCert
	}

	return res, diags
}

func (m *endpointMssqlSSLMethodEncryptedVerifyCert) convert() (*endpoint_airbyte.MSSQLSource_SSLConfig_EncryptedVerifyCertificate, diag.Diagnostics) {
	return &endpoint_airbyte.MSSQLSource_SSLConfig_EncryptedVerifyCertificate{
		EncryptedVerifyCertificate: &endpoint_airbyte.MSSQLSource_SSLEncryptedVerifyCert{
			HostNameInCertificate: m.HostNameInCertificate.ValueString(),
		},
	}, nil
}

func (m *endpointMssqlSourceSettings) parse(e *endpoint_airbyte.MSSQLSource) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Host = types.StringValue(e.GetHost())
	m.Port = types.Int64Value(e.GetPort())
	m.Database = types.StringValue(e.GetDatabase())
	m.Username = types.StringValue(e.GetUsername())

	if e.ReplicationMethod != endpoint_airbyte.MSSQLSource_MSSQL_REPLICATION_METHOD_UNSPECIFIED {
		m.ReplicationMethod = types.StringValue(e.ReplicationMethod.String())
	}

	if e.SslMethod != nil {
		if m.SSLMethod == nil {
			m.SSLMethod = new(endpointMssqlSSLMethod)
		}
		diags.Append(m.SSLMethod.parse(e.GetSslMethod())...)
	}

	return diags
}

func (m *endpointMssqlSSLMethod) parse(e *endpoint_airbyte.MSSQLSource_SSLConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case e.GetUnencrypted() != nil:
		m.Unencrypted = new(endpointMssqlSSLMethodUnencrypted)
	case e.GetEncryptedTrustServerCertificate() != nil:
		m.EncryptedTrusted = new(endpointMssqlSSLMethodEncryptedTrusted)
	case e.GetEncryptedVerifyCertificate() != nil:
		if m.EncryptedVerifyCert == nil {
			m.EncryptedVerifyCert = new(endpointMssqlSSLMethodEncryptedVerifyCert)
		}
		diags.Append(m.EncryptedVerifyCert.parse(e.GetEncryptedVerifyCertificate())...)
	}

	return diags
}

func (m *endpointMssqlSSLMethodEncryptedVerifyCert) parse(e *endpoint_airbyte.MSSQLSource_SSLEncryptedVerifyCert) diag.Diagnostics {
	m.HostNameInCertificate = types.StringValue(e.GetHostNameInCertificate())

	return nil
}

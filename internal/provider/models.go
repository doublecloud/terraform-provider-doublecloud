package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	v1 "github.com/doublecloud/go-genproto/doublecloud/v1"
)

type AccessModel struct {
	Ipv4CIDRBlocks []*CIDRBlock   `tfsdk:"ipv4_cidr_blocks"`
	Ipv6CIDRBlocks []*CIDRBlock   `tfsdk:"ipv6_cidr_blocks"`
	DataServices   []types.String `tfsdk:"data_services"`
}

type CIDRBlock struct {
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
}

func accessDataServiceOneofValues() []string {
	result := make([]string, 0)
	for k, v := range v1.Access_DataService_value {
		if v == 0 {
			continue
		}
		result = append(result, strings.ToLower(strings.TrimPrefix(k, "DATA_SERVICE_")))
	}
	return result
}

func (m *AccessModel) convert() (*v1.Access, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := v1.Access{}
	if m.DataServices != nil {
		r.DataServices = &v1.Access_DataServiceList{}
		r.DataServices.Values = make([]v1.Access_DataService, len(m.DataServices))
		for i := 0; i < len(m.DataServices); i++ {
			k := fmt.Sprintf("DATA_SERVICE_%v", strings.ToUpper(m.DataServices[i].ValueString()))
			r.DataServices.Values[i] = v1.Access_DataService(v1.Access_DataService_value[k])
		}
	}
	if m.Ipv4CIDRBlocks != nil {
		r.Ipv4CidrBlocks = &v1.Access_CidrBlockList{}
		r.Ipv4CidrBlocks.Values = make([]*v1.Access_CidrBlock, len(m.Ipv4CIDRBlocks))
		for i := 0; i < len(m.Ipv4CIDRBlocks); i++ {
			v, d := m.Ipv4CIDRBlocks[i].convert()
			r.Ipv4CidrBlocks.Values[i] = v
			diags.Append(d...)
		}
	}
	if m.Ipv6CIDRBlocks != nil {
		r.Ipv6CidrBlocks = &v1.Access_CidrBlockList{}
		r.Ipv6CidrBlocks.Values = make([]*v1.Access_CidrBlock, len(m.Ipv6CIDRBlocks))
		for i := 0; i < len(m.Ipv6CIDRBlocks); i++ {
			v, d := m.Ipv6CIDRBlocks[i].convert()
			r.Ipv6CidrBlocks.Values[i] = v
			diags.Append(d...)
		}
	}
	return &r, diags
}

func (m *AccessModel) parse(t *v1.Access) diag.Diagnostics {
	var diags diag.Diagnostics

	if services := t.GetDataServices(); services != nil {
		if values := services.GetValues(); values != nil {
			if m.DataServices == nil {
				m.DataServices = make([]types.String, 0)
			}
			for i := 0; i < len(values); i++ {
				if i >= len(m.DataServices) {
					service := strings.ToLower(strings.TrimPrefix(values[i].String(), "DATA_SERVICE_"))
					m.DataServices = append(m.DataServices, types.StringValue(service))
				}
			}
		} else {
			m.DataServices = nil
		}
	} else {
		m.DataServices = nil
	}
	if ipv4 := t.GetIpv4CidrBlocks(); ipv4 != nil {
		if values := ipv4.GetValues(); values != nil {
			if m.Ipv4CIDRBlocks == nil {
				m.Ipv4CIDRBlocks = make([]*CIDRBlock, 0)
			}
			for i, block := range values {
				if i >= len(m.Ipv4CIDRBlocks) {
					m.Ipv4CIDRBlocks = append(m.Ipv4CIDRBlocks, new(CIDRBlock))
				}
				diags.Append(m.Ipv4CIDRBlocks[i].parse(block)...)
			}
		} else {
			m.Ipv4CIDRBlocks = nil
		}
	} else {
		m.Ipv4CIDRBlocks = nil
	}
	if ipv6 := t.GetIpv6CidrBlocks(); ipv6 != nil {
		if values := ipv6.GetValues(); values != nil {
			if m.Ipv6CIDRBlocks == nil {
				m.Ipv6CIDRBlocks = make([]*CIDRBlock, 0)
			}
			for i, block := range values {
				if i >= len(m.Ipv6CIDRBlocks) {
					m.Ipv6CIDRBlocks = append(m.Ipv6CIDRBlocks, new(CIDRBlock))
				}
				diags.Append(m.Ipv6CIDRBlocks[i].parse(block)...)
			}
		} else {
			m.Ipv6CIDRBlocks = nil
		}
	} else {
		m.Ipv6CIDRBlocks = nil
	}

	return diags
}

func (m *CIDRBlock) convert() (*v1.Access_CidrBlock, diag.Diagnostics) {
	return &v1.Access_CidrBlock{
		Value:       m.Value.ValueString(),
		Description: m.Description.ValueString(),
	}, nil
}

func (m *CIDRBlock) parse(v *v1.Access_CidrBlock) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Value = types.StringValue(v.GetValue())
	m.Description = types.StringValue(v.GetDescription())

	return diags
}

func AccessSchemaBlock() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"data_services": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "List of allowed services",
				ElementType:         types.StringType,
				Validators:          []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf(accessDataServiceOneofValues()...))},
			},
			"ipv4_cidr_blocks": schema.ListNestedAttribute{
				Optional:            true,
				NestedObject:        CIDRBlockAttributeSchema(),
				MarkdownDescription: "IPv4 CIDR blocks",
			},
			"ipv6_cidr_blocks": schema.ListNestedAttribute{
				Optional:            true,
				NestedObject:        CIDRBlockAttributeSchema(),
				MarkdownDescription: "IPv6 CIDR blocks",
			},
		},
	}
}

func CIDRBlockAttributeSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "CIDR block",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "CIDR block description",
			},
		},
	}
}

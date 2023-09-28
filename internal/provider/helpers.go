package provider

import (
	"fmt"

	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// convertSchemaAttributes helps to convert resource schema to datasource schema.
// All attributes marked as Computed, not Required and not Optional.
// You can modify any of the attributes later.
func convertSchemaAttributes(resAttrs map[string]resourceschema.Attribute, dataAttrs map[string]dataschema.Attribute) diag.Diagnostics {
	var diags diag.Diagnostics

	for name, attrInterface := range resAttrs {
		switch attr := attrInterface.(type) {
		case resourceschema.StringAttribute:
			dataAttrs[name] = convertStringAttribute(attr)
		case resourceschema.SingleNestedAttribute:
			dataAttrs[name] = convertSingleNestedAttribute(attr, diags)
		default:
			diags.AddError("can not convert resource attribute to datasource attribute", fmt.Sprintf("unsupported type for attribute %q: %v", name, attr))
		}
	}

	return diags
}

func convertStringAttribute(attr resourceschema.StringAttribute) *dataschema.StringAttribute {
	return &dataschema.StringAttribute{
		Computed:            true,
		Sensitive:           attr.Sensitive,
		Description:         attr.Description,
		MarkdownDescription: attr.MarkdownDescription,
		DeprecationMessage:  attr.DeprecationMessage,
	}
}

func convertSingleNestedAttribute(attr resourceschema.SingleNestedAttribute, diags diag.Diagnostics) *dataschema.SingleNestedAttribute {
	dataAttrs := make(map[string]dataschema.Attribute)

	diags.Append(convertSchemaAttributes(attr.Attributes, dataAttrs)...)
	return &dataschema.SingleNestedAttribute{
		Attributes:          dataAttrs,
		Computed:            true,
		Sensitive:           attr.Sensitive,
		Description:         attr.Description,
		MarkdownDescription: attr.MarkdownDescription,
		DeprecationMessage:  attr.DeprecationMessage,
	}
}

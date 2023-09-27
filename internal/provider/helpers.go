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
func convertSchemaAttributes(attrs map[string]resourceschema.Attribute, diagn diag.Diagnostics) map[string]dataschema.Attribute {
	res := make(map[string]dataschema.Attribute)

	for name, attrInterface := range attrs {
		switch attr := attrInterface.(type) {
		case resourceschema.StringAttribute:
			res[name] = convertStringAttribute(attr)
		case resourceschema.SingleNestedAttribute:
			res[name] = convertSingleNestedAttribute(attr, diagn)
		default:
			diagn.AddError("can not convert resource attribute to datasource attribute", fmt.Sprintf("unsupported type for attribute %q: %v", name, attr))
		}
	}

	return res
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

func convertSingleNestedAttribute(attr resourceschema.SingleNestedAttribute, diagn diag.Diagnostics) *dataschema.SingleNestedAttribute {
	return &dataschema.SingleNestedAttribute{
		Attributes:          convertSchemaAttributes(attr.Attributes, diagn),
		Computed:            true,
		Sensitive:           attr.Sensitive,
		Description:         attr.Description,
		MarkdownDescription: attr.MarkdownDescription,
		DeprecationMessage:  attr.DeprecationMessage,
	}
}

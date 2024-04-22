package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func protoEnumValidator(keys map[int32]string) validator.String {
	names := make([]string, len(keys))
	for i, v := range keys {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
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

type suppressAutoscaledDiskDiff struct{}

var _ planmodifier.Int64 = &suppressAutoscaledDiskDiff{}

func (*suppressAutoscaledDiskDiff) Description(context.Context) string {
	return "suppress diff if disk size was autoscaled"
}

func (s *suppressAutoscaledDiskDiff) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (*suppressAutoscaledDiskDiff) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, rsp *planmodifier.Int64Response) {
	// Ignore if it's creation/deletion or no diff
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() || req.PlanValue.Equal(req.StateValue) {
		return
	}
	// Ignore scale up
	if req.PlanValue.ValueInt64() > req.StateValue.ValueInt64() {
		return
	}

	var maxSize types.Int64
	diag := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("max_disk_size"), &maxSize)
	if diag.HasError() {
		rsp.Diagnostics.Append(diag...)
		return
	}

	// Autoscaling disabled, intentional disk scale down, only recreation possible.
	if maxSize.IsNull() {
		rsp.RequiresReplace = true
		return
	}

	if req.StateValue.ValueInt64() <= maxSize.ValueInt64() {
		rsp.Diagnostics.AddWarning(
			"disk size was autoscaled",
			fmt.Sprintf("Disk size at path %s was autoscaled, ignoring changes."+
				"\nTo remove that warning set value of %s to %d", req.Path.String(), req.Path.String(), req.StateValue.ValueInt64()))
		rsp.PlanValue = req.StateValue
	} else {
		rsp.RequiresReplace = true
	}
}

type clusterResourcesValidator struct{}

func (*clusterResourcesValidator) Description(context.Context) string {
	return "validate resource configuration"
}

func (v *clusterResourcesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

var _ validator.Object = &clusterResourcesValidator{}

func (*clusterResourcesValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, rsp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() {
		return
	}

	presetPresent := !req.ConfigValue.Attributes()["resource_preset_id"].IsNull()
	minPresetPresent := !req.ConfigValue.Attributes()["min_resource_preset_id"].IsNull()
	maxPresetPresent := !req.ConfigValue.Attributes()["max_resource_preset_id"].IsNull()

	if presetPresent && minPresetPresent {
		rsp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			`Attribute "resource_preset_id" cannot be specified when "min_resource_preset_id" is specified`,
		))
		return
	}
	if presetPresent && maxPresetPresent {
		rsp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			`Attribute "resource_preset_id" cannot be specified when "max_resource_preset_id" is specified`,
		))
		return
	}

	if minPresetPresent != maxPresetPresent {
		rsp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			`Attribute "min_resource_preset_id" must be specified when "max_resource_preset_id" is specified`,
		))
		return
	}

	if !presetPresent && !minPresetPresent {
		rsp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			`At least one attribute out of [resource_preset_id, (min_resource_preset_id, max_resource_preset_id)] must be specified`,
		))
		return
	}
}

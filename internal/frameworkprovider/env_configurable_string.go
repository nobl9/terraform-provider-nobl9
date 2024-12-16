package frameworkprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringValuable = envConfigurableString{}
	_ basetypes.StringTypable  = envConfigurableStringType{}
)

// envConfigurableString is a custom [basetypes.StringValuable] that works with [envconfig].
// Example:
//
//	type MyType struct {
//		Value envConfigurableString `envconfig:"MY_VALUE"`
//	}
type envConfigurableString struct{ basetypes.StringValue }

// Decode implements [envconfig.Decoder] interface.
func (e *envConfigurableString) Decode(value string) error {
	e.StringValue = basetypes.NewStringValue(value)
	return nil
}

// envConfigurableStringType is a custom [basetypes.StringTypable] which accompanies [envConfigurableString].
// In order for the [envConfigurableString] to work, you need to set this type on the attribute.
// Example:
type envConfigurableStringType struct{ basetypes.StringType }

// ValueFromTerraform overrides the default [basetypes.StringType.ValueFromTerraform] method.
// It returns [envConfigurableString] instead of a [basetypes.StringValue].
func (t envConfigurableStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	v, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	sv, ok := v.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("expected %T to return a %T", t.StringType, basetypes.StringValue{})
	}
	return envConfigurableString{sv}, nil
}

package frameworkprovider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.BoolValuable = envConfigurableBool{}
	_ basetypes.BoolTypable  = envConfigurableBoolType{}
)

type envConfigurableBool struct{ basetypes.BoolValue }

// Decode implements [envconfig.Decoder] interface.
func (e *envConfigurableBool) Decode(value string) error {
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	e.BoolValue = basetypes.NewBoolValue(boolValue)
	return nil
}

type envConfigurableBoolType struct{ basetypes.BoolType }

func (t envConfigurableBoolType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	v, err := t.BoolType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	sv, ok := v.(basetypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("expected %T to return a %T", t.BoolType, basetypes.BoolValue{})
	}
	return envConfigurableBool{sv}, nil
}

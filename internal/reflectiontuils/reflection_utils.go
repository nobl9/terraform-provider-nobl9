package reflectiontuils

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// GetAttributeTypes uses reflection to extract tfsdk struct tags from any struct
// and constructs a map[string][attr.Type] by calling [attr.Value.Type] method on each field.
// This is useful for automatically generating attribute type maps for Terraform Framework models.
func GetAttributeTypes(structValue any) map[string]attr.Type {
	val := reflect.ValueOf(structValue)
	typ := reflect.TypeOf(structValue)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val = reflect.New(typ.Elem()).Elem()
			typ = typ.Elem()
		} else {
			val = val.Elem()
			typ = typ.Elem()
		}
	}

	attributeTypes := make(map[string]attr.Type)
	if typ.Kind() != reflect.Struct {
		return attributeTypes
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		if !field.IsExported() {
			continue
		}
		tfsdkTag := field.Tag.Get("tfsdk")
		if tfsdkTag == "" || tfsdkTag == "-" {
			continue
		}
		typeMethod := fieldValue.MethodByName("Type")
		if !typeMethod.IsValid() {
			continue
		}
		nilContext := reflect.Zero(reflect.TypeOf((*context.Context)(nil)).Elem())
		results := typeMethod.Call([]reflect.Value{nilContext})
		if len(results) == 1 {
			if attrType, ok := results[0].Interface().(attr.Type); ok {
				attributeTypes[tfsdkTag] = attrType
			}
		}
	}
	return attributeTypes
}

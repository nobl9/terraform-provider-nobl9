package nobl9

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestAgentSchemaFieldCompleteness verifies that all fields defined in nobl9-go
// agent config structs are present in the corresponding Terraform schema.
// This test helps catch schema drift when the nobl9-go library adds new fields.
func TestAgentSchemaFieldCompleteness(t *testing.T) {
	fullSchema := agentSchema()

	for _, agentConfig := range getSupportedAgentConfigs() {
		t.Run(agentConfig.configKey, func(t *testing.T) {
			// Get the schema for this agent config
			configSchema, ok := fullSchema[agentConfig.configKey]
			if !ok {
				t.Errorf("Schema key %q not found in agent schema", agentConfig.configKey)
				return
			}

			// Get fields from nobl9-go struct
			structFields := getStructJSONFields(agentConfig.configStruct)

			// Get fields from Terraform schema
			schemaFields := getSchemaFields(configSchema)
			schemaFieldSet := make(map[string]bool)
			for _, f := range schemaFields {
				schemaFieldSet[f] = true
			}

			// Check for missing fields
			var missingFields []string
			for _, jsonField := range structFields {
				schemaField := toSchemaFieldName(jsonField)
				if !schemaFieldSet[schemaField] {
					missingFields = append(missingFields, jsonField+" (expected: "+schemaField+")")
				}
			}

			if len(missingFields) > 0 {
				t.Errorf("Schema %q is missing fields present in nobl9-go struct %T:\n  %s",
					agentConfig.configKey,
					agentConfig.configStruct,
					strings.Join(missingFields, "\n  "))
			}
		})
	}
}

// getJSONFieldName extracts the JSON field name from a struct field's json tag.
// It handles tags like `json:"field_name,omitempty"` and returns "field_name".
func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return ""
	}
	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

// getStructJSONFields returns all JSON field names for a given struct.
// Handles both pointer and value types.
func getStructJSONFields(configStruct any) []string {
	var fields []string
	t := reflect.TypeOf(configStruct)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonName := getJSONFieldName(field)
		if jsonName != "" {
			fields = append(fields, jsonName)
		}
	}
	return fields
}

// getSchemaFields returns all field names from a Terraform schema's nested resource.
func getSchemaFields(s *schema.Schema) (fields []string) {
	if s.Elem == nil {
		return fields
	}
	resource, ok := s.Elem.(*schema.Resource)
	if !ok {
		return fields
	}
	fields = make([]string, 0, len(resource.Schema))
	for fieldName := range resource.Schema {
		fields = append(fields, fieldName)
	}
	return fields
}

// toSchemaFieldName converts a JSON field name (camelCase) to Terraform schema
// field name (snake_case).
func toSchemaFieldName(jsonName string) string {
	var result strings.Builder
	for i, r := range jsonName {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32) // Convert to lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

package nobl9

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
)

func TestElasticsearchDirectSpec(t *testing.T) {
	provider := elasticsearchDirectSpec{}
	resourceSchema := provider.GetSchema()

	assert.True(t, resourceSchema["url"].Required)
	assert.True(t, resourceSchema["api_key"].Required)
	assert.True(t, resourceSchema["api_key"].Sensitive)

	spec := provider.MarshalSpec(mockResourceData{
		"url":     "https://example.aws.found.io",
		"api_key": "encoded-api-key",
	})
	require.NotNil(t, spec.Elasticsearch)
	assert.Equal(t, "https://example.aws.found.io", spec.Elasticsearch.URL)
	assert.Equal(t, "encoded-api-key", spec.Elasticsearch.APIKey)
}

func TestElasticsearchDirectSpecPreservesAPIKeyOnRead(t *testing.T) {
	provider := elasticsearchDirectSpec{}
	resourceSchema := provider.GetSchema()
	resourceSchema["description"] = schemaDescription()
	data := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"url":         "https://old.aws.found.io",
		"api_key":     "encoded-api-key",
		"description": "",
	})

	diags := provider.UnmarshalSpec(data, v1alphaDirect.Spec{
		Description: "updated",
		Elasticsearch: &v1alphaDirect.ElasticsearchConfig{
			URL:    "https://new.aws.found.io",
			APIKey: "[hidden]",
		},
	})

	require.False(t, diags.HasError())
	assert.Equal(t, "https://new.aws.found.io", data.Get("url"))
	assert.Equal(t, "encoded-api-key", data.Get("api_key"))
	assert.Equal(t, "updated", data.Get("description"))
}

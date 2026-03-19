package nobl9

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

type mockResourceData map[string]any

func (m mockResourceData) Get(key string) any {
	return m[key]
}

func (m mockResourceData) GetOk(key string) (any, bool) {
	v, ok := m[key]
	return v, ok
}

func (m mockResourceData) GetRawConfig() cty.Value {
	return cty.NilVal
}

func TestValidateServiceNowAuth(t *testing.T) {
	for name, tc := range map[string]struct {
		resourceData mockResourceData
		expectError  bool
	}{
		"basic auth valid": {
			resourceData: mockResourceData{
				"username":  "user",
				"password":  "pass",
				"api_token": "",
			},
			expectError: false,
		},
		"token auth valid": {
			resourceData: mockResourceData{
				"username":  "",
				"password":  "",
				"api_token": "token",
			},
			expectError: false,
		},
		"basic auth missing password": {
			resourceData: mockResourceData{
				"username":  "user",
				"password":  "",
				"api_token": "",
			},
			expectError: true,
		},
		"token auth missing token": {
			resourceData: mockResourceData{
				"username":  "",
				"password":  "",
				"api_token": "",
			},
			expectError: true,
		},
		"both auth methods provided": {
			resourceData: mockResourceData{
				"username":  "user",
				"password":  "pass",
				"api_token": "token",
			},
			expectError: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := validateServiceNowAuth(tc.resourceData)
			if tc.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestAlertMethodServiceNowMarshalSpec(t *testing.T) {
	provider := alertMethodServiceNow{}

	basicSpec := provider.MarshalSpec(mockResourceData{
		"description":     "desc",
		"username":        "user",
		"password":        "pass",
		"api_token":       "",
		"instance_name":   "instance",
		"send_resolution": nil,
	})
	if basicSpec.ServiceNow == nil {
		t.Fatal("expected servicenow spec")
	}
	if basicSpec.ServiceNow.Username != "user" || basicSpec.ServiceNow.Password != "pass" {
		t.Fatal("expected basic credentials in payload")
	}
	if basicSpec.ServiceNow.ApiToken != "" {
		t.Fatal("expected api token to be empty for basic auth")
	}

	tokenSpec := provider.MarshalSpec(mockResourceData{
		"description":     "desc",
		"username":        "",
		"password":        "",
		"api_token":       "token",
		"instance_name":   "instance",
		"send_resolution": nil,
	})
	if tokenSpec.ServiceNow == nil {
		t.Fatal("expected servicenow spec")
	}
	if tokenSpec.ServiceNow.ApiToken != "token" {
		t.Fatal("expected api token in payload")
	}
	if tokenSpec.ServiceNow.Username != "" || tokenSpec.ServiceNow.Password != "" {
		t.Fatal("expected basic credentials to be empty for token auth")
	}
}

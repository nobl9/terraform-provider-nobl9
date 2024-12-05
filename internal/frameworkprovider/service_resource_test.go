package frameworkprovider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/stretchr/testify/assert"
)

func TestAccExampleResource(t *testing.T) {
	serviceName := fmt.Sprintf("service-%d", time.Now().UnixNano())

	serviceResource := serviceResourceTemplateModel{
		ResourceName:         "test",
		ServiceResourceModel: getExampleServiceResource(),
	}
	serviceResource.ServiceResourceModel.Labels = appendTestLabels(serviceResource.ServiceResourceModel.Labels)
	serviceResource.ServiceResourceModel.Name = serviceName

	manifestService := v1alphaService.New(
		v1alphaService.Metadata{
			Name:        serviceName,
			DisplayName: "Service",
			Project:     "default",
			Annotations: v1alpha.MetadataAnnotations{"key": "value"},
			Labels: v1alpha.Labels{
				"team":   []string{"green"},
				"env":    []string{"dev", "prod"},
				"origin": []string{"terraform-acc-test"},
			},
		},
		v1alphaService.Spec{
			Description: "Example service",
		},
	)
	manifestService.Status = &v1alphaService.Status{
		SloCount: 0,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: newServiceResource(t, serviceResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasApplied(t, manifestService),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionCreate),
					},
				},
			},
			//// ImportState testing.
			//{
			//	ResourceName:      "nobl9_service.test",
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//	// This is not normally necessary, but is here because this
			//	// example code does not have an actual upstream service.
			//	// Once the Read method is able to refresh information from
			//	// the upstream service, this can be removed.
			//	ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			//},
			// Update and Read testing.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.ServiceResourceModel.DisplayName = types.StringValue("New Service Display Name")
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "display_name", "New Service Display Name"),
					assertResourceWasApplied(t, func() v1alphaService.Service {
						svc := manifestService
						svc.Metadata.DisplayName = "New Service Display Name"
						return svc
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestRenderServiceResourceTemplate(t *testing.T) {
	t.Parallel()

	actual := newServiceResource(t, serviceResourceTemplateModel{
		ResourceName:         "this",
		ServiceResourceModel: getExampleServiceResource(),
	})

	expected := `resource "nobl9_service" "this" {
  name = "service"
  display_name = "Service"
  project = "default"
  annotations = {
    key = "value",
  }
  label {
    key = "team"
    values = [
      "green",
    ]
  }
  label {
    key = "env"
    values = [
      "prod",
      "dev",
    ]
  }
  description = "Example service"
}
`

	assert.Equal(t, expected, actual)
}

type serviceResourceTemplateModel struct {
	ResourceName string
	ServiceResourceModel
}

func newServiceResource(t *testing.T, model serviceResourceTemplateModel) string {
	return executeTemplate(t, "service_resource.hcl.tmpl", model)
}

func getExampleServiceResource() ServiceResourceModel {
	return ServiceResourceModel{
		Name:        "service",
		DisplayName: types.StringValue("Service"),
		Project:     "default",
		Description: types.StringValue("Example service"),
		Annotations: map[string]string{"key": "value"},
		Labels: Labels{
			{Key: "team", Values: []string{"green"}},
			{Key: "env", Values: []string{"prod", "dev"}},
		},
	}
}

package frameworkprovider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/stretchr/testify/assert"
)

func TestAccServiceResource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	unixNow := time.Now().UnixNano()
	serviceName := fmt.Sprintf("service-%d", unixNow)
	serviceNameRecreatedByNameChange := fmt.Sprintf("service-name-recreated-%d", unixNow)

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
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: newServiceResource(t, serviceResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasApplied(t, ctx, manifestService),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// Delete.
			{
				Config: newServiceResource(t, serviceResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasDeleted(t, ctx, manifestService),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionDestroy),
					},
				},
				Destroy: true,
			},
			// ImportState - invalid id.
			{
				ResourceName:  fmt.Sprintf("nobl9_service.test_%d", unixNow),
				ImportStateId: serviceName,
				ImportState:   true,
				ExpectError:   regexp.MustCompile(`Invalid import ID`),
			},
			// ImportState.
			{
				ResourceName:  "nobl9_service.test",
				ImportStateId: "default/" + serviceName,
				ImportState:   true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if !assert.Len(t, states, 1) {
						return errors.New("expected exactly one state")
					}
					assert.Equal(t, serviceName, states[0].Attributes["name"])
					assert.Equal(t, "default", states[0].Attributes["project"])
					return nil
				},
				// In the next step we're also verifying the imported state, so we need to persist it.
				ImportStatePersist: true,
				PreConfig:          func() { applyNobl9Objects(t, ctx, manifestService) },
			},
			// Update and Read, ensure computed field does not pollute the plan.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.ServiceResourceModel.DisplayName = types.StringValue("New Service Display Name")
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "display_name", "New Service Display Name"),
					assertResourceWasApplied(t, ctx, func() v1alphaService.Service {
						svc := manifestService
						svc.Metadata.DisplayName = "New Service Display Name"
						return svc
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectNoChangeInPlan{attrName: "status"},
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Update name - recreate.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.ServiceResourceModel.Name = serviceNameRecreatedByNameChange
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "name", serviceNameRecreatedByNameChange),
					assertResourceWasApplied(t, ctx, func() v1alphaService.Service {
						svc := manifestService
						svc.Metadata.Name = serviceNameRecreatedByNameChange
						return svc
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionReplace),
					},
				},
			},
			// Update project - recreate.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.ServiceResourceModel.Name = serviceNameRecreatedByNameChange
					m.ServiceResourceModel.Project = "default-recreated"
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "name", serviceNameRecreatedByNameChange),
					resource.TestCheckResourceAttr("nobl9_service.test", "project", "default-recreated"),
					assertResourceWasApplied(t, ctx, func() v1alphaService.Service {
						svc := manifestService
						svc.Metadata.Name = serviceNameRecreatedByNameChange
						svc.Metadata.Project = "default-recreated"
						return svc
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionReplace),
					},
				},
			},
			// Delete automatically occurs in TestCase, no need to clean up.
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

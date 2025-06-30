package frameworkprovider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/stretchr/testify/assert"
)

func TestAccServiceResource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()

	auxiliaryObjects := []manifest.Object{manifestProject}

	serviceNameRecreatedByNameChange := generateName()
	serviceResource := serviceResourceTemplateModel{
		ResourceName:         "test",
		ServiceResourceModel: getExampleServiceResource(t),
	}
	serviceResource.Project = manifestProject.GetName()

	manifestService := serviceResource.ToManifest()
	manifestService.Status = &v1alphaService.Status{
		SloCount: 0,
	}

	res := newServiceResource(t, serviceResource)
	res = res
	recreatedProjectName := generateName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				PreConfig: func() {
					applyNobl9Objects(t, ctx, auxiliaryObjects...)
					t.Cleanup(func() { deleteNobl9Objects(t, ctx, auxiliaryObjects...) })
				},
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
				ResourceName:  "nobl9_service.test",
				ImportStateId: serviceResource.Name,
				ImportState:   true,
				ExpectError:   regexp.MustCompile(`Invalid import ID`),
			},
			// ImportState.
			{
				ResourceName:  "nobl9_service.test",
				ImportStateId: manifestProject.GetName() + "/" + serviceResource.Name,
				ImportState:   true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if !assert.Len(t, states, 1) {
						return errors.New("expected exactly one state")
					}
					assert.Equal(t, serviceResource.Name, states[0].Attributes["name"])
					assert.Equal(t, serviceResource.Project, states[0].Attributes["project"])
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
					m.DisplayName = types.StringValue("New Service Display Name")
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
					m.Name = serviceNameRecreatedByNameChange
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
					m.Name = serviceNameRecreatedByNameChange
					m.Project = recreatedProjectName
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "name", serviceNameRecreatedByNameChange),
					resource.TestCheckResourceAttr("nobl9_service.test", "project", recreatedProjectName),
					assertResourceWasApplied(t, ctx, func() v1alphaService.Service {
						svc := manifestService
						svc.Metadata.Name = serviceNameRecreatedByNameChange
						svc.Metadata.Project = recreatedProjectName
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

	exampleService := getExampleServiceResource(t)
	exampleService.Name = "service"
	actual := newServiceResource(t, serviceResourceTemplateModel{
		ResourceName:         "this",
		ServiceResourceModel: exampleService,
	})

	expected := fmt.Sprintf(`resource "nobl9_service" "this" {
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
      "dev",
      "prod",
    ]
  }
  label {
    key = "origin"
    values = [
      "terraform-acc-test",
    ]
  }
  label {
    key = "terraform-acc-test-id"
    values = [
      "%d",
    ]
  }
  label {
    key = "terraform-test-name"
    values = [
      "%s",
    ]
  }
  description = "Example service"
}
`, testStartTime.UnixNano(), t.Name())

	assert.Equal(t, expected, actual)
}

type serviceResourceTemplateModel struct {
	ResourceName string
	ServiceResourceModel
}

func newServiceResource(t *testing.T, model serviceResourceTemplateModel) string {
	return executeTemplate(t, "service_resource.hcl.tmpl", model)
}

func getExampleServiceResource(t *testing.T) ServiceResourceModel {
	return ServiceResourceModel{
		Name:        generateName(),
		DisplayName: types.StringValue("Service"),
		Project:     "default",
		Description: types.StringValue("Example service"),
		Annotations: map[string]string{"key": "value"},
		Labels: annotateLabels(t, Labels{
			{Key: "team", Values: []string{"green"}},
			{Key: "env", Values: []string{"dev", "prod"}},
		}),
	}
}

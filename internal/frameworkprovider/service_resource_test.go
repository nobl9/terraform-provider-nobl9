package frameworkprovider

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
	"github.com/stretchr/testify/assert"
)

func TestAccServiceResource(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()

	auxiliaryObjects := []manifest.Object{manifestProject}

	serviceNameRecreatedByNameChange := e2etestutils.GenerateName()
	serviceResource := serviceResourceTemplateModel{
		ResourceName:         "test",
		ServiceResourceModel: getExampleServiceResource(t),
	}
	serviceResource.Project = manifestProject.GetName()

	manifestService := serviceResource.ToManifest()
	manifestService.Status = &v1alphaService.Status{
		SloCount: 0,
	}

	recreatedProjectName := e2etestutils.GenerateName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create and Read.
			{
				PreConfig: func() {
					e2etestutils.V1Apply(t, auxiliaryObjects)
					t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })
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
			// 2. Delete.
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
			// 3. ImportState - invalid id.
			{
				ResourceName:  "nobl9_service.test",
				ImportStateId: serviceResource.Name,
				ImportState:   true,
				ExpectError:   regexp.MustCompile(`Invalid import ID`),
			},
			// 4. ImportState.
			{
				ResourceName:  "nobl9_service.test",
				ImportStateId: serviceResource.Project + "/" + serviceResource.Name,
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
				PreConfig:          func() { e2etestutils.V1Apply(t, []manifest.Object{manifestService}) },
			},
			// 5. Update and Read, ensure computed field does not pollute the plan.
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
						expectChangesInResourcePlan(planDiff{Modified: []string{"display_name"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// 6. Update name and revert display name - recreate.
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
						expectChangesInResourcePlan(planDiff{
							Modified: []string{"name", "display_name"},
							Removed:  []string{"status"},
						}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionReplace),
					},
				},
			},
			// 7. Update project - recreate.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.Name = serviceNameRecreatedByNameChange
					m.Project = recreatedProjectName
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
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
						expectChangesInResourcePlan(planDiff{
							Modified: []string{"project"},
							Removed:  []string{"status"},
						}),
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

	exampleResource := getExampleServiceResource(t)
	exampleResource.Name = "service"
	exampleResource.Labels = Labels{
		{Key: "team", Values: []string{"green", "orange"}},
		{Key: "env", Values: []string{"prod"}},
		{Key: "empty", Values: []string{""}},
	}
	actual := newServiceResource(t, serviceResourceTemplateModel{
		ResourceName:         "this",
		ServiceResourceModel: exampleResource,
	})

	assertHCL(t, actual)
	assert.Equal(t, readExpectedConfig(t, "service-config.tf"), actual)
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
		Name:        e2etestutils.GenerateName(),
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

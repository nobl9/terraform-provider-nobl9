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
					assertResourceWasApplied(t, ctx, serviceResource.ToManifest()),
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
					assertResourceWasDeleted(t, ctx, serviceResource.ToManifest()),
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
				PreConfig:          func() { e2etestutils.V1Apply(t, []manifest.Object{serviceResource.ToManifest()}) },
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
						svc := serviceResource.ToManifest()
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
						svc := serviceResource.ToManifest()
						svc.Metadata.Name = serviceNameRecreatedByNameChange
						return svc
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{
							Modified: []string{"name", "display_name"},
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
						svc := serviceResource.ToManifest()
						svc.Metadata.Name = serviceNameRecreatedByNameChange
						svc.Metadata.Project = recreatedProjectName
						return svc
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{
							Modified: []string{"project"},
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

func TestAccServiceResource_planValidation(t *testing.T) {
	t.Parallel()
	testAccSetup(t)

	serviceResource := serviceResourceTemplateModel{
		ResourceName:         "test",
		ServiceResourceModel: getExampleServiceResource(t),
	}
	serviceResource.Name = "not valid"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: newServiceResource(t, serviceResource),
				ExpectError: regexp.MustCompile(
					`(?m)Bad Request: Validation for Service 'not valid' in project 'default' has\nfailed`,
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccServiceResource_ResponsibleUsers(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()
	auxiliaryObjects := []manifest.Object{manifestProject}

	serviceResource := serviceResourceTemplateModel{
		ResourceName:         "test",
		ServiceResourceModel: getExampleServiceResource(t),
	}
	serviceResource.Project = manifestProject.GetName()
	serviceResource.ResponsibleUsers = []ResponsibleUserModel{
		{ID: types.StringValue("user1@example.com")},
		{ID: types.StringValue("user2@example.com")},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create service with responsible users.
			{
				PreConfig: func() {
					e2etestutils.V1Apply(t, auxiliaryObjects)
					t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })
				},
				Config: newServiceResource(t, serviceResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.#", "2"),
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.0.id", "user1@example.com"),
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.1.id", "user2@example.com"),
					assertResourceWasApplied(t, ctx, serviceResource.ToManifest()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// 2. Update responsible users - add one more.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.ResponsibleUsers = []ResponsibleUserModel{
						{ID: types.StringValue("user1@example.com")},
						{ID: types.StringValue("user2@example.com")},
						{ID: types.StringValue("user3@example.com")},
					}
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.#", "3"),
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.0.id", "user1@example.com"),
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.1.id", "user2@example.com"),
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.2.id", "user3@example.com"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"responsible_users"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// 3. Remove all responsible users.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.ResponsibleUsers = nil
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"responsible_users"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccServiceResource_ReviewCycle(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()
	auxiliaryObjects := []manifest.Object{manifestProject}

	serviceResource := serviceResourceTemplateModel{
		ResourceName:         "test",
		ServiceResourceModel: getExampleServiceResource(t),
	}
	serviceResource.Project = manifestProject.GetName()
	serviceResource.ReviewCycle = &ReviewCycleModel{
		RRule:     types.StringValue("FREQ=WEEKLY;BYDAY=MO"),
		StartTime: types.StringValue("2024-01-01T09:00:00"),
		TimeZone:  types.StringValue("America/New_York"),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create service with review cycle.
			{
				PreConfig: func() {
					e2etestutils.V1Apply(t, auxiliaryObjects)
					t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })
				},
				Config: newServiceResource(t, serviceResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.rrule", "FREQ=WEEKLY;BYDAY=MO"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.start_time", "2024-01-01T09:00:00"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.time_zone", "America/New_York"),
					assertResourceWasApplied(t, ctx, serviceResource.ToManifest()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// 2. Update review cycle - change time zone and rrule.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.ReviewCycle = &ReviewCycleModel{
						RRule:     types.StringValue("FREQ=MONTHLY;BYMONTHDAY=1"),
						StartTime: types.StringValue("2024-01-01T09:00:00"),
						TimeZone:  types.StringValue("Europe/London"),
					}
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.rrule", "FREQ=MONTHLY;BYMONTHDAY=1"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.start_time", "2024-01-01T09:00:00"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.time_zone", "Europe/London"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"review_cycle"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// 3. Remove review cycle.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.ReviewCycle = nil
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("nobl9_service.test", "review_cycle.rrule"),
					resource.TestCheckNoResourceAttr("nobl9_service.test", "review_cycle.start_time"),
					resource.TestCheckNoResourceAttr("nobl9_service.test", "review_cycle.time_zone"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"review_cycle"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccServiceResource_ResponsibleUsersAndReviewCycle(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()
	auxiliaryObjects := []manifest.Object{manifestProject}

	serviceResource := serviceResourceTemplateModel{
		ResourceName:         "test",
		ServiceResourceModel: getExampleServiceResource(t),
	}
	serviceResource.Project = manifestProject.GetName()
	serviceResource.ResponsibleUsers = []ResponsibleUserModel{
		{ID: types.StringValue("user1@example.com")},
	}
	serviceResource.ReviewCycle = &ReviewCycleModel{
		RRule:     types.StringValue("FREQ=WEEKLY;BYDAY=FR"),
		StartTime: types.StringValue("2024-01-05T14:00:00"),
		TimeZone:  types.StringValue("UTC"),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create service with both responsible users and review cycle.
			{
				PreConfig: func() {
					e2etestutils.V1Apply(t, auxiliaryObjects)
					t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })
				},
				Config: newServiceResource(t, serviceResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.#", "1"),
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.0.id", "user1@example.com"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.rrule", "FREQ=WEEKLY;BYDAY=FR"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.start_time", "2024-01-05T14:00:00"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.time_zone", "UTC"),
					assertResourceWasApplied(t, ctx, serviceResource.ToManifest()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// 2. Update both responsible users and review cycle.
			{
				Config: newServiceResource(t, func() serviceResourceTemplateModel {
					m := serviceResource
					m.ResponsibleUsers = []ResponsibleUserModel{
						{ID: types.StringValue("user2@example.com")},
						{ID: types.StringValue("user3@example.com")},
					}
					m.ReviewCycle = &ReviewCycleModel{
						RRule:     types.StringValue("FREQ=DAILY"),
						StartTime: types.StringValue("2024-01-01T08:00:00"),
						TimeZone:  types.StringValue("Asia/Tokyo"),
					}
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.#", "2"),
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.0.id", "user2@example.com"),
					resource.TestCheckResourceAttr("nobl9_service.test", "responsible_users.1.id", "user3@example.com"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.rrule", "FREQ=DAILY"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.start_time", "2024-01-01T08:00:00"),
					resource.TestCheckResourceAttr("nobl9_service.test", "review_cycle.time_zone", "Asia/Tokyo"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"responsible_users", "review_cycle"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.test", plancheck.ResourceActionUpdate),
					},
				},
			},
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
		Labels: addTestLabels(t, Labels{
			{Key: "team", Values: []string{"green"}},
			{Key: "env", Values: []string{"dev", "prod"}},
		}),
		ReviewCycle: &ReviewCycleModel{
			RRule:     types.StringValue("FREQ=DAILY"),
			StartTime: types.StringValue("2024-01-01T08:00:00"),
			TimeZone:  types.StringValue("Asia/Tokyo"),
		},
		ResponsibleUsers: []ResponsibleUserModel{
			{ID: types.StringValue("userID1")},
			{ID: types.StringValue("userID2")},
		},
	}
}

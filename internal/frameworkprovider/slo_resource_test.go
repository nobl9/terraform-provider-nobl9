package frameworkprovider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaExamples "github.com/nobl9/nobl9-go/manifest/v1alpha/examples"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccSLOResource(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()
	manifestService := getExampleServiceResource(t).ToManifest()
	manifestService.Metadata.Project = manifestProject.GetName()
	auxiliaryObjects := []manifest.Object{manifestProject, manifestService}

	manifestDirect := e2etestutils.ProvisionStaticDirect(t, v1alpha.AppDynamics)

	manifestAlertPolicy := e2etestutils.GetExampleObject[v1alphaAlertPolicy.AlertPolicy](
		t,
		manifest.KindAlertPolicy,
		nil,
	)
	manifestAlertPolicy.Metadata.Name = e2etestutils.GenerateName()
	manifestAlertPolicy.Metadata.Project = manifestProject.GetName()
	manifestAlertPolicy.Metadata.Labels = e2etestutils.AnnotateLabels(t, nil)
	manifestAlertPolicy.Spec.AlertMethods = nil
	auxiliaryObjects = append(auxiliaryObjects, manifestAlertPolicy)

	sloNameRecreatedByNameChange := e2etestutils.GenerateName()
	sloResource := sloResourceTemplateModel{
		ResourceName:     "test",
		SLOResourceModel: getExampleSLOResource(t),
	}
	sloResource.Name = e2etestutils.GenerateName()
	sloResource.Project = manifestProject.GetName()
	sloResource.Service = manifestService.GetName()
	sloResource.Indicator = []IndicatorModel{{
		Name:    manifestDirect.GetName(),
		Project: types.StringValue(manifestDirect.GetProject()),
		Kind:    types.StringValue(manifestDirect.GetKind().String()),
	}}

	manifestSLO := sloResource.ToManifest()

	sloConfig := newSLOResource(t, sloResource)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create and Read.
			{
				PreConfig: func() {
					e2etestutils.V1Apply(t, auxiliaryObjects)
					t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })
				},
				Config: sloConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasApplied(t, ctx, manifestSLO),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// 2. Delete.
			{
				Config: sloConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasDeleted(t, ctx, manifestSLO),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionDestroy),
					},
				},
				Destroy: true,
			},
			// 3. ImportState - invalid id.
			{
				ResourceName:  "nobl9_slo.test",
				ImportStateId: sloResource.Name,
				ImportState:   true,
				ExpectError:   regexp.MustCompile(`Invalid import ID`),
			},
			// 4. ImportState.
			{
				ResourceName:  "nobl9_slo.test",
				Config:        sloConfig,
				ImportStateId: sloResource.Project + "/" + sloResource.Name,
				ImportState:   true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if !assert.Len(t, states, 1) {
						return errors.New("expected exactly one state")
					}
					assert.Equal(t, sloResource.Name, states[0].Attributes["name"])
					assert.Equal(t, sloResource.Project, states[0].Attributes["project"])
					return nil
				},
				// In the next step we're also verifying the imported state, so we need to persist it.
				ImportStatePersist: true,
				PreConfig: func() {
					e2etestutils.V1Apply(t, []manifest.Object{manifestSLO})
				},
			},
			// 5. Update and Read, ensure computed fields do not pollute the plan.
			{
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.DisplayName = types.StringValue("New SLO Display Name")
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_slo.test", "display_name", "New SLO Display Name"),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO
						slo.Metadata.DisplayName = "New SLO Display Name"
						return slo
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"display_name"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// 6. Update name and revert display name - recreate.
			{
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.Name = sloNameRecreatedByNameChange
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_slo.test", "name", sloNameRecreatedByNameChange),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO
						slo.Metadata.Name = sloNameRecreatedByNameChange
						return slo
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"name", "display_name"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionReplace),
					},
				},
			},
			// Delete automatically occurs in TestCase, no need to clean up.
		},
	})
}

func TestAccSLOResource_moveSLO(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()
	manifestService := getExampleServiceResource(t).ToManifest()
	manifestService.Metadata.Project = manifestProject.GetName()
	auxiliaryObjects := []manifest.Object{manifestProject, manifestService}

	manifestDirect := e2etestutils.ProvisionStaticDirect(t, v1alpha.AppDynamics)

	manifestAlertPolicy := e2etestutils.GetExampleObject[v1alphaAlertPolicy.AlertPolicy](
		t,
		manifest.KindAlertPolicy,
		nil,
	)
	manifestAlertPolicy.Metadata.Name = e2etestutils.GenerateName()
	manifestAlertPolicy.Metadata.Project = manifestProject.GetName()
	manifestAlertPolicy.Metadata.Labels = e2etestutils.AnnotateLabels(t, nil)
	manifestAlertPolicy.Spec.AlertMethods = nil
	auxiliaryObjects = append(auxiliaryObjects, manifestAlertPolicy)

	sloResource := sloResourceTemplateModel{
		ResourceName:     "test",
		SLOResourceModel: getExampleSLOResource(t),
	}
	sloResource.Name = e2etestutils.GenerateName()
	sloResource.Project = manifestProject.GetName()
	sloResource.Service = manifestService.GetName()
	sloResource.Indicator = []IndicatorModel{{
		Name:    manifestDirect.GetName(),
		Project: types.StringValue(manifestDirect.GetProject()),
		Kind:    types.StringValue(manifestDirect.GetKind().String()),
	}}
	sloResource.AlertPolicies = []string{manifestAlertPolicy.GetName()}

	manifestSLO := sloResource.ToManifest()

	newProjectName := e2etestutils.GenerateName()
	newServiceName := e2etestutils.GenerateName()

	sloConfig := newSLOResource(t, sloResource)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create and Read.
			{
				PreConfig: func() {
					e2etestutils.V1Apply(t, auxiliaryObjects)
					t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })
				},
				Config: sloConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasApplied(t, ctx, manifestSLO),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// 2. Update project with alert policies - error.
			{
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.Project = "new-project"
					return m
				}()),
				ExpectError: regexp.MustCompile(`Cannot move SLO between Projects with attached Alert Policies.`),
			},
			// 3. Remove Alert Policy from SLO.
			{
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.AlertPolicies = nil
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("nobl9_slo.test", "alert_policies"),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO
						slo.Spec.AlertPolicies = nil
						return slo
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"alert_policies"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// 4. Update project and another attribute - error.
			{
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.Project = "new-project"
					m.AlertPolicies = nil
					m.DisplayName = stringValue("Changed display!")
					return m
				}()),
				ExpectError: regexp.MustCompile(
					"When changing the `project`, no other attribute can be modified, except for `service`."),
			},
			// 5. Update project - move SLO.
			{
				PreConfig: func() {
					newProjectManifest := manifestProject
					newProjectManifest.Metadata.Name = newProjectName
					newServiceManifest := manifestService
					newServiceManifest.Metadata.Project = newProjectName

					t.Cleanup(func() {
						e2etestutils.V1Delete(t, []manifest.Object{newProjectManifest, newServiceManifest})
					})
				},
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.AlertPolicies = nil
					m.Project = newProjectName
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_slo.test", "project", newProjectName),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO
						slo.Spec.AlertPolicies = nil
						slo.Metadata.Project = newProjectName
						return slo
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"project"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// 6. Update project and service - move SLO (back to the original Project).
			{
				PreConfig: func() {
					newServiceManifest := manifestService
					newServiceManifest.Metadata.Project = manifestProject.GetName()
					newServiceManifest.Metadata.Name = newServiceName

					t.Cleanup(func() {
						e2etestutils.V1Delete(t, []manifest.Object{newServiceManifest})
					})
				},
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.AlertPolicies = nil
					m.Project = manifestProject.GetName()
					m.Service = newServiceName
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_slo.test", "project", manifestProject.GetName()),
					resource.TestCheckResourceAttr("nobl9_slo.test", "service", newServiceName),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO
						slo.Spec.AlertPolicies = nil
						slo.Metadata.Project = manifestProject.GetName()
						slo.Spec.Service = newServiceName
						return slo
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"project", "service"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Delete automatically occurs in TestCase, no need to clean up.
		},
	})
}

func TestAccSLOResource_moveTwoSLOs(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()
	manifestService := getExampleServiceResource(t).ToManifest()
	manifestService.Metadata.Project = manifestProject.GetName()
	auxiliaryObjects := []manifest.Object{manifestProject, manifestService}

	manifestDirect := e2etestutils.ProvisionStaticDirect(t, v1alpha.AppDynamics)

	sloResource1 := sloResourceTemplateModel{
		ResourceName:     "first",
		SLOResourceModel: getExampleSLOResource(t),
	}
	sloResource1.Name = e2etestutils.GenerateName()
	sloResource1.Project = manifestProject.GetName()
	sloResource1.Service = manifestService.GetName()
	sloResource1.Indicator = []IndicatorModel{{
		Name:    manifestDirect.GetName(),
		Project: types.StringValue(manifestDirect.GetProject()),
		Kind:    types.StringValue(manifestDirect.GetKind().String()),
	}}
	sloResource1.AlertPolicies = nil
	sloResource2 := sloResource1
	sloResource2.Name = e2etestutils.GenerateName()
	sloResource2.ResourceName = "second"

	manifestSLO1 := sloResource1.ToManifest()
	manifestSLO2 := sloResource2.ToManifest()

	newProjectName := e2etestutils.GenerateName()

	sloConfig1 := newSLOResource(t, sloResource1)
	sloConfig2 := newSLOResource(t, sloResource2)

	combinedConfig := sloConfig1 + "\n" + sloConfig2

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create and Read.
			{
				PreConfig: func() {
					e2etestutils.V1Apply(t, auxiliaryObjects)
					t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })
				},
				Config: combinedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasApplied(t, ctx, manifestSLO1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.first", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction("nobl9_slo.second", plancheck.ResourceActionCreate),
					},
				},
			},
			// 2. Update project - move SLOs.
			{
				PreConfig: func() {
					newProjectManifest := manifestProject
					newProjectManifest.Metadata.Name = newProjectName
					newServiceManifest := manifestService
					newServiceManifest.Metadata.Project = newProjectName

					t.Cleanup(func() {
						e2etestutils.V1Delete(t, []manifest.Object{newProjectManifest, newServiceManifest})
					})
				},
				Config: func() string {
					m1 := sloResource1
					m1.Project = newProjectName
					m2 := sloResource2
					m2.Project = newProjectName
					return newSLOResource(t, m1) + "\n" + newSLOResource(t, m2)
				}(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_slo.first", "project", newProjectName),
					resource.TestCheckResourceAttr("nobl9_slo.second", "project", newProjectName),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO1
						slo.Metadata.Project = newProjectName
						return slo
					}()),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO2
						slo.Metadata.Project = newProjectName
						return slo
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcesPlan(map[string]planDiff{
							"nobl9_slo.first":  {Modified: []string{"project"}},
							"nobl9_slo.second": {Modified: []string{"project"}},
						}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.first", plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("nobl9_slo.second", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Delete automatically occurs in TestCase, no need to clean up.
		},
	})
}

func TestAccSLOResource_moveDeprecatedCompositeV1SLO(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()
	manifestService := getExampleServiceResource(t).ToManifest()
	manifestService.Metadata.Project = manifestProject.GetName()
	auxiliaryObjects := []manifest.Object{manifestProject, manifestService}

	manifestDirect := e2etestutils.ProvisionStaticDirect(t, v1alpha.AppDynamics)

	sloResource := sloResourceTemplateModel{
		ResourceName:     "test",
		SLOResourceModel: getExampleSLOResource(t),
	}
	sloResource.Name = e2etestutils.GenerateName()
	sloResource.Project = manifestProject.GetName()
	sloResource.Service = manifestService.GetName()
	sloResource.Indicator = []IndicatorModel{{
		Name:    manifestDirect.GetName(),
		Project: types.StringValue(manifestDirect.GetProject()),
		Kind:    types.StringValue(manifestDirect.GetKind().String()),
	}}
	sloResource.AlertPolicies = nil
	sloResource.Composite = []CompositeV1Model{{
		Target: types.Float64Value(0.5),
		BurnRateCondition: []CompositeV1BurnRateConditionModel{{
			Op:    types.StringValue("gt"),
			Value: types.Float64Value(1),
		}},
	}}

	manifestSLO := sloResource.ToManifest()

	newProjectName := e2etestutils.GenerateName()

	sloConfig := newSLOResource(t, sloResource)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. Create and Read.
			{
				PreConfig: func() {
					e2etestutils.V1Apply(t, auxiliaryObjects)
					t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })
				},
				Config: sloConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasApplied(t, ctx, manifestSLO),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// 2. Update project - move SLO.
			{
				PreConfig: func() {
					newProjectManifest := manifestProject
					newProjectManifest.Metadata.Name = newProjectName
					newServiceManifest := manifestService
					newServiceManifest.Metadata.Project = newProjectName

					t.Cleanup(func() {
						e2etestutils.V1Delete(t, []manifest.Object{newProjectManifest, newServiceManifest})
					})
				},
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.AlertPolicies = nil
					m.Project = newProjectName
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_slo.test", "project", newProjectName),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO
						slo.Spec.AlertPolicies = nil
						slo.Metadata.Project = newProjectName
						return slo
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"project"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Delete automatically occurs in TestCase, no need to clean up.
		},
	})
}

func TestAccSLOResource_customScenario(t *testing.T) {
	t.Parallel()
	testAccSetup(t)

	sloConfig := `resource "nobl9_slo" "this" {
  name             = "test-composite-tf-new"
  service          = "test-tf"
  budgeting_method = "Timeslices"
  project          = "test-tf"

  attachment {
    url          = "https://test/"
    display_name = "!#@#$#%^&%^&*(*(()&^%$%;:900897hhnkxz'dsdklsjkhjssjk😂☝️"
  }

  time_window {
    unit       = "Day"
    count      = 1
    is_rolling = true
  }

  objective {
    display_name = "AA"
    # name = "tf-objective-1"
    target       = 0.98
    time_slice_target = 0.9
    # value        = 0
    composite {
      max_delay = "1m"
      components {
        objectives {
          composite_objective {
            project      = "test-permissions-18-06"
            slo          = "test-raw-with-composite"
            objective    = "a"
            weight       = 1.0
            when_delayed = "CountAsGood"
          }
          composite_objective {
            project      = "test-permissions-18-06"
            slo          = "test-ratio"
            objective    = "existing-good-and-total"
            weight       = 1.0
            when_delayed = "CountAsBad"
          }
           composite_objective {
            project      = nobl9_slo.test-ratio-tf.project
            slo          = "test-ratio-tf"
            objective    = "tf-objective-1"
            weight       = 1.0
            when_delayed = "CountAsGood"
          }
          composite_objective {
            project      = nobl9_slo.test-raw-tf.project
            slo          = nobl9_slo.test-raw-tf.name
            objective    = "objective-1"
            weight       = 1.0
            when_delayed = "Ignore"
          }
        }
      }
    }
  }

  alert_policies = [
    # "test-tf-alert-1", "test-tf-alert-2"
  ]

  # indicator {
  #   name    = "web-prometheus123456"
  #   kind    = "Agent"
  #   project = "default"
  # }
}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: sloConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.this", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccSLOResource_custom(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()
	manifestService := getExampleServiceResource(t).ToManifest()
	manifestService.Metadata.Project = manifestProject.GetName()

	auxiliaryObjects := []manifest.Object{
		manifestProject,
		manifestService,
	}

	manifestAlertPolicy1 := e2etestutils.GetExampleObject[v1alphaAlertPolicy.AlertPolicy](
		t,
		manifest.KindAlertPolicy,
		nil,
	)
	manifestAlertPolicy1.Metadata.Name = e2etestutils.GenerateName()
	manifestAlertPolicy1.Metadata.Project = manifestProject.GetName()
	manifestAlertPolicy1.Metadata.Labels = e2etestutils.AnnotateLabels(t, nil)
	manifestAlertPolicy1.Spec.AlertMethods = nil
	manifestAlertPolicy2 := manifestAlertPolicy1
	manifestAlertPolicy2.Metadata.Name = e2etestutils.GenerateName()
	auxiliaryObjects = append(auxiliaryObjects, manifestAlertPolicy1, manifestAlertPolicy2)

	manifestAlertMethod := e2etestutils.GetExampleObject[v1alphaAlertMethod.AlertMethod](
		t,
		manifest.KindAlertMethod,
		e2etestutils.FilterExamplesByAlertMethodType(v1alpha.AlertMethodTypeEmail),
	)
	manifestAlertMethod.Metadata.Name = e2etestutils.GenerateName()
	manifestAlertMethod.Metadata.Project = manifestProject.GetName()
	auxiliaryObjects = append(auxiliaryObjects, manifestAlertMethod)

	e2etestutils.V1Apply(t, auxiliaryObjects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })

	tests := map[string]struct {
		sloResourceModelModifier func(t *testing.T, model SLOResourceModel) SLOResourceModel
		sloManifestModifier      func(t *testing.T, model v1alphaSLO.SLO) v1alphaSLO.SLO
		expectedError            string
	}{
		"with empty alert policies": {
			sloResourceModelModifier: func(t *testing.T, model SLOResourceModel) SLOResourceModel {
				model.AlertPolicies = []string{}
				return model
			},
			expectedError: "Attribute alert_policies set must contain at least 1 elements, got: 0",
		},
		"with alert policies": {
			sloResourceModelModifier: func(t *testing.T, model SLOResourceModel) SLOResourceModel {
				model.AlertPolicies = []string{manifestAlertPolicy1.GetName(), manifestAlertPolicy2.GetName()}
				return model
			},
		},
		"with anomaly config": {
			sloResourceModelModifier: func(t *testing.T, model SLOResourceModel) SLOResourceModel {
				model.AnomalyConfig = []AnomalyConfigModel{{
					NoData: []AnomalyConfigNoDataModel{{
						AlertAfter: stringValue("1h"),
						AlertMethods: []AnomalyConfigAlertMethodModel{{
							Name:    manifestAlertMethod.GetName(),
							Project: manifestAlertMethod.GetProject(),
						}},
					}},
				}}
				return model
			},
		},
		"no display name": {
			sloResourceModelModifier: func(t *testing.T, model SLOResourceModel) SLOResourceModel {
				model.DisplayName = types.StringNull()
				return model
			},
		},
		"empty composite block": {
			sloResourceModelModifier: func(t *testing.T, model SLOResourceModel) SLOResourceModel {
				slo := e2etestutils.GetExampleObject[v1alphaSLO.SLO](
					t,
					manifest.KindSLO,
					func(example v1alphaExamples.Example) bool {
						return strings.Contains(example.GetVariant(), "composite")
					},
				)
				compositeModel := newSLOResourceConfigFromManifest(slo)
				compositeModel.Objectives = compositeModel.Objectives[:1]
				compositeModel.Objectives[0].Composite = []CompositeObjectiveModel{{
					MaxDelay:   types.StringValue("15m"),
					Components: nil,
				}}
				model.Objectives = compositeModel.Objectives
				model.Indicator = nil
				return model
			},
		},
		"ratio metric with no value in objective": {
			sloResourceModelModifier: func(t *testing.T, model SLOResourceModel) SLOResourceModel {
				model.Objectives[0].RawMetric = nil
				model.Objectives[0].Value = types.Float64Value(1)
				model.Objectives[0].Op = types.String{}
				model.Objectives[0].CountMetrics = []CountMetricsModel{{
					Incremental: types.BoolValue(false),
					Good: []MetricSpecModel{{
						Prometheus: []PrometheusModel{{
							PromQL: "sum(rate(http_request_duration_seconds_count{job=\"api\"}[5m]))",
						}},
					}},
					Total: []MetricSpecModel{{
						Prometheus: []PrometheusModel{{
							PromQL: "sum(rate(http_request_duration_seconds_count{job=\"api\"}[5m]))",
						}},
					}},
				}}
				return model
			},
			sloManifestModifier: func(t *testing.T, slo v1alphaSLO.SLO) v1alphaSLO.SLO {
				slo.Spec.Objectives[0].Value = ptr(1.0)
				return slo
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sloModel := getExampleSLOResource(t)
			sloModel.Name = e2etestutils.GenerateName()
			sloModel.Project = manifestProject.GetName()
			sloModel.Service = manifestService.GetName()
			sloModel = test.sloResourceModelModifier(t, sloModel)

			manifestSLO := sloModel.ToManifest()

			if !manifestSLO.Spec.HasCompositeObjectives() {
				typ := manifestSLO.Spec.AllMetricSpecs()[0].DataSourceType()
				var dataSource manifest.Object
				switch sloModel.Indicator[0].Kind.ValueString() {
				case manifest.KindDirect.String():
					dataSource = e2etestutils.ProvisionStaticDirect(t, typ)
				default:
					dataSource = e2etestutils.ProvisionStaticAgent(t, typ)
				}
				sloModel.Indicator[0].Name = dataSource.GetName()
				sloModel.Indicator[0].Project = types.StringValue(
					dataSource.(manifest.ProjectScopedObject).GetProject(),
				)
			}

			manifestSLO = sloModel.ToManifest()
			if test.sloManifestModifier != nil {
				manifestSLO = test.sloManifestModifier(t, manifestSLO)
			}

			sloConfig := newSLOResource(t, sloResourceTemplateModel{
				ResourceName:     "test",
				SLOResourceModel: sloModel,
			})

			if test.expectedError != "" {
				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						// Create and Read.
						{
							Config:      sloConfig,
							ExpectError: regexp.MustCompile(test.expectedError),
							PlanOnly:    true,
						},
					},
				})
				return
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read.
					{
						Config: sloConfig,
						Check: resource.ComposeAggregateTestCheckFunc(
							assertResourceWasApplied(t, ctx, manifestSLO),
						),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectNonEmptyPlan(),
								plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionCreate),
							},
						},
					},
					// Delete.
					{
						Config: sloConfig,
						Check: resource.ComposeAggregateTestCheckFunc(
							assertResourceWasDeleted(t, ctx, manifestSLO),
						),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectNonEmptyPlan(),
								plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionDestroy),
							},
						},
						Destroy: true,
					},
				},
			})
		})
	}
}

func TestAccSLOResource_objectiveValueErrors(t *testing.T) {
	t.Parallel()
	testAccSetup(t)

	tests := map[string]struct {
		configFunc    func() string
		expectedError string
	}{
		"composite with value": {
			configFunc: func() string {
				slo := e2etestutils.GetExampleObject[v1alphaSLO.SLO](
					t,
					manifest.KindSLO,
					func(example v1alphaExamples.Example) bool {
						return strings.Contains(example.GetVariant(), "composite")
					},
				)
				model := newSLOResourceConfigFromManifest(slo)
				for i, objective := range model.Objectives {
					if len(objective.Composite) > 0 {
						model.Objectives[i].Value = types.Float64Value(0.0)
						break
					}
				}
				sloConfig := newSLOResource(t, sloResourceTemplateModel{
					ResourceName:     "this",
					SLOResourceModel: *model,
				})
				return sloConfig
			},
			expectedError: "objective value cannot be set when defining composite SLOs",
		},
		"threshold without value": {
			configFunc: func() string {
				slo := e2etestutils.GetExampleObject[v1alphaSLO.SLO](
					t,
					manifest.KindSLO,
					func(example v1alphaExamples.Example) bool {
						slo := example.GetObject().(v1alphaSLO.SLO)
						return slo.Spec.HasRawMetric()
					},
				)
				for i := range slo.Spec.Objectives {
					slo.Spec.Objectives[i].Value = nil
				}
				sloConfig := newSLOResource(t, sloResourceTemplateModel{
					ResourceName:     "this",
					SLOResourceModel: *newSLOResourceConfigFromManifest(slo),
				})
				return sloConfig
			},
			expectedError: "objective value must be set for ratio and threshold objectives",
		},
		"ratio without value": {
			configFunc: func() string {
				slo := e2etestutils.GetExampleObject[v1alphaSLO.SLO](
					t,
					manifest.KindSLO,
					func(example v1alphaExamples.Example) bool {
						slo := example.GetObject().(v1alphaSLO.SLO)
						return slo.Spec.HasCountMetrics()
					},
				)
				for i := range slo.Spec.Objectives {
					slo.Spec.Objectives[i].Value = nil
				}
				sloConfig := newSLOResource(t, sloResourceTemplateModel{
					ResourceName:     "this",
					SLOResourceModel: *newSLOResourceConfigFromManifest(slo),
				})
				return sloConfig
			},
			expectedError: "objective value must be set for ratio and threshold objectives",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read.
					{
						Config:      test.configFunc(),
						ExpectError: regexp.MustCompile(test.expectedError),
						PlanOnly:    true,
					},
				},
			})
		})
	}
}

const slosPerService = 50

// nolint: gocognit
func TestAccSLOResource_examples(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()

	auxiliaryObjects := []manifest.Object{manifestProject}

	sloExamples := e2etestutils.GetAllExamples(t, manifest.KindSLO)
	// Composite SLOs depend on other SLOs. Example SLOs are being sorted so that Composite SLOs are placed at the end,
	// allowing them to depend on the SLOs listed before them.
	slices.SortStableFunc(sloExamples, func(i, j v1alphaExamples.Example) int {
		var intI, intJ int
		iSlo := i.GetObject().(v1alphaSLO.SLO)
		if iSlo.Spec.HasCompositeObjectives() {
			intI = 1
		}
		jSlo := j.GetObject().(v1alphaSLO.SLO)
		if jSlo.Spec.HasCompositeObjectives() {
			intJ = 1
		}
		return intI - intJ
	})

	type testCase struct {
		example v1alphaExamples.Example
		slo     v1alphaSLO.SLO
	}

	testCases := make([]testCase, 0, len(sloExamples))
	var service v1alphaService.Service
	for i, example := range sloExamples {
		if example.GetVariant() == "generic" {
			continue
		}

		slo := example.GetObject().(v1alphaSLO.SLO)
		slo.Metadata = v1alphaSLO.Metadata{
			Name:        e2etestutils.GenerateName(),
			DisplayName: fmt.Sprintf("SLO %d", i),
			Project:     manifestProject.GetName(),
			Labels:      e2etestutils.AnnotateLabels(t, nil),
			Annotations: commonAnnotations,
		}
		// Generate new service for every `slosPerService` SLOs to meet the quota.
		if i%slosPerService == 0 {
			service = v1alphaService.New(
				v1alphaService.Metadata{
					Name:    e2etestutils.GenerateName(),
					Project: manifestProject.GetName(),
				},
				v1alphaService.Spec{
					Description: e2etestutils.GetObjectDescription(),
				},
			)
			auxiliaryObjects = append(auxiliaryObjects, service)
		}
		slo.Spec.Service = service.GetName()
		slo.Spec.AlertPolicies = nil
		slo.Spec.AnomalyConfig = nil

		if slo.Spec.HasCompositeObjectives() {
			for componentIndex, component := range slo.Spec.Objectives[0].Composite.Objectives {
				componentSlo := testCases[len(testCases)-1-componentIndex].slo
				componentSlo.Metadata.Name = e2etestutils.GenerateName()
				component.Project = componentSlo.GetProject()
				component.SLO = componentSlo.GetName()
				component.Objective = componentSlo.Spec.Objectives[0].Name
				auxiliaryObjects = append(auxiliaryObjects, componentSlo)
				slo.Spec.Objectives[0].Composite.Objectives[componentIndex] = component
			}
		} else {
			metricSpecs := slo.Spec.AllMetricSpecs()
			require.Greater(t, len(metricSpecs), 0, "expected at least 1 metric spec")

			sourceType := metricSpecs[0].DataSourceType()
			var source manifest.Object
			switch slo.Spec.Indicator.MetricSource.Kind {
			case manifest.KindDirect:
				source = e2etestutils.ProvisionStaticDirect(t, sourceType)
			default:
				source = e2etestutils.ProvisionStaticAgent(t, sourceType)
			}
			slo.Spec.Indicator.MetricSource.Name = source.GetName()
			slo.Spec.Indicator.MetricSource.Project = source.(manifest.ProjectScopedObject).GetProject()

			// TODO: Remove this after PC-13575 is resolved.
			if slo.Spec.Indicator.MetricSource.Kind == manifest.KindAgent && sourceType == v1alpha.CloudWatch {
				skip := false
				for _, spec := range slo.Spec.AllMetricSpecs() {
					if spec.CloudWatch.AccountID != nil {
						skip = true
						break
					}
				}
				if skip {
					continue
				}
			}
		}
		testCases = append(testCases, testCase{
			example: example,
			slo:     slo,
		})
	}

	e2etestutils.V1Apply(t, auxiliaryObjects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })

	for _, tc := range testCases {
		t.Run(testNameFromExample(tc.example), func(t *testing.T) {
			t.Parallel()

			sloConfig := newSLOResource(t, sloResourceTemplateModel{
				ResourceName:     "test",
				SLOResourceModel: *newSLOResourceConfigFromManifest(tc.slo),
			})

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// Create and Read.
					{
						Config: sloConfig,
						Check: resource.ComposeAggregateTestCheckFunc(
							assertResourceWasApplied(t, ctx, tc.slo),
						),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectNonEmptyPlan(),
								plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionCreate),
							},
						},
					},
					// Delete.
					{
						Config: sloConfig,
						Check: resource.ComposeAggregateTestCheckFunc(
							assertResourceWasDeleted(t, ctx, tc.slo),
						),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectNonEmptyPlan(),
								plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionDestroy),
							},
						},
						Destroy: true,
					},
				},
			})
		})
	}
}

func TestRenderSLOResourceTemplate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		expectedFile     string
		resourceModifier func(model SLOResourceModel) SLOResourceModel
	}{
		"config": {
			expectedFile: "slo-config.tf",
			resourceModifier: func(model SLOResourceModel) SLOResourceModel {
				model.AlertPolicies = []string{"alert-policy"}
				model.Labels = Labels{
					{Key: "team", Values: []string{"green", "orange"}},
					{Key: "env", Values: []string{"prod"}},
					{Key: "empty", Values: []string{""}},
				}
				return model
			},
		},
		"nested objects in metric spec": {
			expectedFile: "slo-nested-objects-in-metric-spec.tf",
			resourceModifier: func(model SLOResourceModel) SLOResourceModel {
				model.AlertPolicies = nil
				model.Labels = nil
				model.Annotations = nil
				model.Objectives[0].RawMetric[0].Query[0] = MetricSpecModel{
					Instana: []InstanaModel{{
						MetricType: "application",
						Application: []InstanaApplicationModel{{
							MetricID:        "some_id",
							Aggregation:     "foo",
							IncludeInternal: types.BoolValue(true),
							GroupBy: []InstanaGroupByModel{{
								Tag: "some-tag",
							}},
						}},
					}},
				}
				return model
			},
		},
		"multiline query": {
			expectedFile: "slo-multiline-query.tf",
			resourceModifier: func(model SLOResourceModel) SLOResourceModel {
				model.AlertPolicies = nil
				model.Labels = nil
				model.Annotations = nil
				model.Objectives[0].RawMetric[0].Query[0] = MetricSpecModel{
					Prometheus: []PrometheusModel{{
						PromQL: `sum by (job) (
  rate(http_request_duration_seconds_count{job="api"}[5m]
)`,
					}},
				}
				return model
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			exampleResource := getExampleSLOResource(t)
			actual := newSLOResource(t, sloResourceTemplateModel{
				ResourceName:     "this",
				SLOResourceModel: test.resourceModifier(exampleResource),
			})

			assertHCL(t, actual)
			assert.Equal(t, readExpectedConfig(t, test.expectedFile), actual)
		})
	}
}

func TestRenderSLOResourceTemplate_examples(t *testing.T) {
	t.Parallel()

	for _, example := range e2etestutils.GetAllExamples(t, manifest.KindSLO) {
		if example.GetVariant() == "generic" {
			continue
		}
		t.Run(testNameFromExample(example), func(t *testing.T) {
			t.Parallel()

			sloManifest := example.GetObject().(v1alphaSLO.SLO)
			resourceModel := newSLOResourceConfigFromManifest(sloManifest)

			config := newSLOResource(t, sloResourceTemplateModel{
				ResourceName:     "this",
				SLOResourceModel: *resourceModel,
			})
			require.True(t, strings.HasPrefix(config, `resource "nobl9_slo" "this" {`),
				`expected config to start with 'resource "nobl9_slo" "this" {'`)

			assertHCL(t, config)
			assert.Equal(t, sloManifest, resourceModel.ToManifest())
		})
	}
}

func TestRenderSLOResourceTemplate_compositeV1Example(t *testing.T) {
	t.Parallel()

	exampleResource := getExampleSLOResource(t)
	exampleResource.AlertPolicies = nil
	exampleResource.Labels = nil
	exampleResource.Annotations = nil
	exampleResource.Indicator = nil
	exampleResource.Objectives = nil

	// Add composite v1 configuration
	exampleResource.Composite = []CompositeV1Model{{
		Target: types.Float64Value(0.95),
		BurnRateCondition: []CompositeV1BurnRateConditionModel{
			{
				Op:    types.StringValue("gt"),
				Value: types.Float64Value(2.0),
			},
			{
				Op:    types.StringValue("lt"),
				Value: types.Float64Value(1.5),
			},
		},
	}}

	actual := newSLOResource(t, sloResourceTemplateModel{
		ResourceName:     "this",
		SLOResourceModel: exampleResource,
	})

	assertHCL(t, actual)
	assert.Equal(t, readExpectedConfig(t, "slo-composite-v1-config.tf"), actual)
}

type sloResourceTemplateModel struct {
	ResourceName string
	SLOResourceModel
}

func newSLOResource(t *testing.T, model sloResourceTemplateModel) string {
	return executeTemplate(t, "slo_resource.hcl.tmpl", model)
}

func getExampleSLOResource(t *testing.T) SLOResourceModel {
	return SLOResourceModel{
		Name:            "slo",
		DisplayName:     types.StringValue("SLO"),
		Project:         "default",
		Description:     types.StringValue("Example SLO"),
		Service:         "service",
		BudgetingMethod: "Occurrences",
		Annotations:     map[string]string{"key": "value"},
		Labels: annotateLabels(t, Labels{
			{Key: "team", Values: []string{"green"}},
		}),
		Indicator: []IndicatorModel{{
			Name:    "indicator",
			Project: types.StringValue("default"),
			Kind:    types.StringValue("Agent"),
		}},
		Objectives: []ObjectiveModel{{
			DisplayName: types.StringValue("obj1"),
			Name:        types.StringValue("tf-objective-1"),
			Op:          types.StringValue("lt"),
			Target:      0.7,
			Value:       types.Float64Value(1),
			RawMetric: []RawMetricModel{{
				Query: []MetricSpecModel{
					{
						AppDynamics: []AppDynamicsModel{{
							ApplicationName: "my_app",
							MetricPath:      "End User Experience|App|End User Response Time 95th percentile (ms)",
						}},
					},
				},
			}},
		}},
		TimeWindow: []TimeWindowModel{{
			Count:     10,
			IsRolling: types.BoolValue(true),
			Unit:      "Minute",
		}},
	}
}

func testNameFromExample(example v1alphaExamples.Example) string {
	name := ""
	if variant := example.GetVariant(); variant != "" {
		name = variant
	}
	if subVariant := example.GetSubVariant(); subVariant != "" {
		name += " - " + subVariant
	}
	return name
}

func ptr[T any](v T) *T { return &v }

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
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
	"github.com/stretchr/testify/assert"
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
	manifestDirect.Metadata.Name = e2etestutils.GenerateName()
	manifestDirect.Metadata.Project = manifestProject.GetName()

	sloNameRecreatedByNameChange := e2etestutils.GenerateName()
	sloResource := sloResourceTemplateModel{
		ResourceName:     "test",
		SLOResourceModel: getExampleSLOResource(t),
	}
	sloResource.Project = manifestProject.GetName()
	sloResource.Service = manifestService.GetName()
	sloResource.Indicator = []IndicatorModel{{
		Name:    manifestDirect.GetName(),
		Project: types.StringValue(manifestDirect.GetProject()),
		Kind:    types.StringValue(manifestDirect.GetKind().String()),
	}}

	manifestSLO := sloResource.ToManifest()

	recreatedProjectName := e2etestutils.GenerateName()

	sloConfig := newSLOResource(t, sloResource)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccSetup(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
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
			// ImportState - invalid id.
			{
				ResourceName:  "nobl9_slo.test",
				ImportStateId: sloResource.Name,
				ImportState:   true,
				ExpectError:   regexp.MustCompile(`Invalid import ID`),
			},
			// ImportState.
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
			// Update and Read, ensure computed field does not pollute the plan.
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
						expectNoChangeInPlan{attrName: "status"},
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Update name - recreate.
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
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionReplace),
					},
				},
			},
			// Update project - recreate.
			{
				PreConfig: func() {
					newProjectManifest := manifestProject
					newProjectManifest.Metadata.Name = recreatedProjectName
					newServiceManifest := manifestService
					newServiceManifest.Metadata.Project = recreatedProjectName

					e2etestutils.V1Apply(t, []manifest.Object{newProjectManifest, newServiceManifest})
					t.Cleanup(func() {
						e2etestutils.V1Delete(t, []manifest.Object{newProjectManifest, newServiceManifest})
					})
				},
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.Project = recreatedProjectName
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_slo.test", "project", recreatedProjectName),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO
						slo.Metadata.Project = recreatedProjectName
						return slo
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionReplace),
					},
				},
			},
			// Delete automatically occurs in TestCase, no need to clean up.
		},
	})
}

func TestAccSLOResource_variants(t *testing.T) {
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

	e2etestutils.V1Apply(t, auxiliaryObjects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, auxiliaryObjects) })

	tests := map[string]struct {
		sloResourceModelModifier func(t *testing.T, model SLOResourceModel) SLOResourceModel
	}{
		"with alert policies": {
			sloResourceModelModifier: func(t *testing.T, model SLOResourceModel) SLOResourceModel {
				model.AlertPolicies = []string{manifestAlertPolicy1.GetName(), manifestAlertPolicy2.GetName()}
				return model
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sloModel := getExampleSLOResource(t)
			sloModel.Project = manifestProject.GetName()
			sloModel.Service = manifestService.GetName()
			sloModel = test.sloResourceModelModifier(t, sloModel)

			manifestSLO := sloModel.ToManifest()
			typ := manifestSLO.Spec.AllMetricSpecs()[0].DataSourceType()
			var dataSource manifest.Object
			switch sloModel.Indicator[0].Kind.ValueString() {
			case manifest.KindDirect.String():
				dataSource = e2etestutils.ProvisionStaticDirect(t, typ)
			default:
				dataSource = e2etestutils.ProvisionStaticAgent(t, typ)
			}
			sloModel.Indicator[0].Name = dataSource.GetName()
			sloModel.Indicator[0].Project = types.StringValue(dataSource.(manifest.ProjectScopedObject).GetProject())
			manifestSLO = sloModel.ToManifest()

			sloConfig := newSLOResource(t, sloResourceTemplateModel{
				ResourceName:     "test",
				SLOResourceModel: sloModel,
			})

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

func TestRenderSLOResourceTemplate(t *testing.T) {
	t.Parallel()

	exampleResource := getExampleSLOResource(t)
	exampleResource.AlertPolicies = []string{"alert-policy"}
	exampleResource.Labels = Labels{
		{Key: "team", Values: []string{"green", "orange"}},
		{Key: "env", Values: []string{"prod"}},
		{Key: "empty", Values: []string{""}},
	}
	actual := newSLOResource(t, sloResourceTemplateModel{
		ResourceName:     "this",
		SLOResourceModel: exampleResource,
	})

	expected := `resource "nobl9_slo" "this" {
  name = "slo"
  display_name = "SLO"
  project = "default"
  annotations = {
    key = "value",
  }
  label {
    key = "team"
    values = [
      "green",
      "orange",
    ]
  }
  label {
    key = "env"
    values = [
      "prod",
    ]
  }
  label {
    key = "empty"
    values = [
      "",
    ]
  }
  description = "Example SLO"

  service = "service"
  budgeting_method = "Occurrences"
  alert_policies = [
    "alert-policy",
  ]

  indicator {
    name = "indicator"
    project = "default"
    kind = "Agent"
  }

  objective {
    display_name = "obj1"
    name = "tf-objective-1"
    op = "lt"
    target = 0.7
    value = 1
    raw_metric {
      query {
        appdynamics {
          application_name = "my_app"
          metric_path = "End User Experience|App|End User Response Time 95th percentile (ms)"
        }
      }
    }
  }

  time_window {
    count = 10
    is_rolling = true
    unit = "Minute"
  }
}
`

	assert.Equal(t, expected, actual)
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
		Objectives: []ObjectiveModel{
			{
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
			},
		},
		TimeWindow: []TimeWindowModel{{
			Count:     10,
			IsRolling: types.BoolValue(true),
			Unit:      "Minute",
		}},
	}
}

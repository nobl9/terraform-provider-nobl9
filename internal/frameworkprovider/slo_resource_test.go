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
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaExamples "github.com/nobl9/nobl9-go/manifest/v1alpha/examples"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccSLOResource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := getExampleProjectResource(t).ToManifest()
	manifestService := getExampleServiceResource(t).ToManifest()
	manifestService.Metadata.Project = manifestProject.GetName()
	manifestDirect := getDirectExampleObject(t, v1alpha.AppDynamics)
	manifestDirect.Metadata.Name = generateName()
	manifestDirect.Metadata.Project = manifestProject.GetName()

	auxiliaryObjects := []manifest.Object{manifestProject, manifestService, manifestDirect}

	sloNameRecreatedByNameChange := generateName()
	sloResource := sloResourceTemplateModel{
		ResourceName:     "test",
		SLOResourceModel: getExampleSLOResource(t),
	}
	sloResource.Project = manifestProject.GetName()
	sloResource.Service = manifestService.GetName()

	manifestSLO := sloResource.ToManifest()

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
				Config: newSLOResource(t, sloResource),
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
				Config: newSLOResource(t, sloResource),
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
				ImportStateId: manifestProject.GetName() + "/" + sloResource.Name,
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
				PreConfig:          func() { applyNobl9Objects(t, ctx, manifestSLO) },
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
				Config: newSLOResource(t, func() sloResourceTemplateModel {
					m := sloResource
					m.Name = sloNameRecreatedByNameChange
					m.Project = recreatedProjectName
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_slo.test", "name", sloNameRecreatedByNameChange),
					resource.TestCheckResourceAttr("nobl9_slo.test", "project", recreatedProjectName),
					assertResourceWasApplied(t, ctx, func() v1alphaSLO.SLO {
						slo := manifestSLO
						slo.Metadata.Name = sloNameRecreatedByNameChange
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

func TestRenderSLOResourceTemplate(t *testing.T) {
	t.Parallel()

	exampleResource := getExampleSLOResource(t)
	exampleResource.AlertPolicies = []string{"alert-policy"}
	actual := newSLOResource(t, sloResourceTemplateModel{
		ResourceName:     "this",
		SLOResourceModel: exampleResource,
	})

	expected := fmt.Sprintf(`resource "nobl9_slo" "this" {
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
  description = "Example SLO"

  service = "service"
  budgeting_method = "Occurrences"
  tier = "1"
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
`, testStartTime.UnixNano(), t.Name())

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
		Tier:            types.StringValue("1"),
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

func getDirectExampleObject(t *testing.T, directType v1alpha.DataSourceType) v1alphaDirect.Direct {
	t.Helper()
	examples := v1alphaExamples.Direct()
	for _, example := range examples {
		direct := example.GetObject().(v1alphaDirect.Direct)
		typ, err := direct.Spec.GetType()
		require.NoError(t, err)
		if typ == directType {
			return direct
		}
	}
	t.Fatalf("could not find direct type %s", directType)
	return v1alphaDirect.Direct{}
}

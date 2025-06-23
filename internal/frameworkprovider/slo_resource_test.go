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
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/stretchr/testify/assert"
)

func TestAccSLOResource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	serviceName := generateName()
	serviceNameRecreatedByNameChange := generateName()

	serviceResource := serviceResourceTemplateModel{
		ResourceName:         "test",
		ServiceResourceModel: getExampleServiceResource(),
	}
	serviceResource.Labels = appendTestLabels(serviceResource.Labels)
	serviceResource.Name = serviceName

	manifestService := v1alphaService.New(
		v1alphaService.Metadata{
			Name:        serviceName,
			DisplayName: "Service",
			Project:     "default",
			Annotations: v1alpha.MetadataAnnotations{"key": "value"},
			Labels:      annotateLabels(t, nil),
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
						plancheck.ExpectResourceAction("nobl9_slo.test", plancheck.ResourceActionCreate),
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
					m.Project = "default-recreated"
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

func TestRenderSLOResourceTemplate(t *testing.T) {
	t.Parallel()

	actual := newSLOResource(t, sloResourceTemplateModel{
		ResourceName:     "this",
		SLOResourceModel: getExampleSLOResource(),
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

func getExampleSLOResource() SLOResourceModel {
	return SLOResourceModel{
		Name:            "slo",
		DisplayName:     types.StringValue("SLO"),
		Project:         "default",
		Description:     types.StringValue("Example SLO"),
		Service:         types.StringValue("service"),
		BudgetingMethod: types.StringValue("Occurrences"),
		Tier:            types.StringValue("1"),
		AlertPolicies:   []string{"alert-policy"},
		Annotations:     map[string]string{"key": "value"},
		Labels: Labels{
			{Key: "team", Values: []string{"green"}},
		},
		Indicator: &IndicatorModel{
			Name:    types.StringValue("indicator"),
			Project: types.StringValue("default"),
			Kind:    types.StringValue("Agent"),
		},
		Objectives: []ObjectiveModel{
			{
				DisplayName: types.StringValue("obj1"),
				Name:        types.StringValue("tf-objective-1"),
				Op:          types.StringValue("lt"),
				Target:      types.Float64Value(0.7),
				Value:       types.Float64Value(1),
				RawMetric: &RawMetricModel{
					Query: []MetricSpecModel{
						{
							AppDynamics: &AppDynamicsModel{
								ApplicationName: types.StringValue("my_app"),
								MetricPath:      types.StringValue("End User Experience|App|End User Response Time 95th percentile (ms)"),
							},
						},
					},
				},
			},
		},
		TimeWindow: &TimeWindowModel{
			Count:     types.Int64Value(10),
			IsRolling: types.BoolValue(true),
			Unit:      types.StringValue("Minute"),
		},
	}
}

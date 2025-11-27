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
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
	"github.com/stretchr/testify/assert"
)

func TestAccProjectResource(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	projectNameRecreatedByNameChange := e2etestutils.GenerateName()
	projectResource := projectResourceTemplateModel{
		ResourceName:         "test",
		ProjectResourceModel: getExampleProjectResource(t),
	}

	manifestProject := projectResource.ToManifest()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: newProjectResource(t, projectResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasApplied(t, ctx, manifestProject),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_project.test", plancheck.ResourceActionCreate),
					},
				},
			},
			// Delete.
			{
				Config: newProjectResource(t, projectResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasDeleted(t, ctx, manifestProject),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_project.test", plancheck.ResourceActionDestroy),
					},
				},
				Destroy: true,
			},
			// ImportState.
			{
				ResourceName:  "nobl9_project.test",
				ImportStateId: projectResource.Name,
				ImportState:   true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if !assert.Len(t, states, 1) {
						return errors.New("expected exactly one state")
					}
					assert.Equal(t, projectResource.Name, states[0].Attributes["name"])
					return nil
				},
				// In the next step we're also verifying the imported state, so we need to persist it.
				ImportStatePersist: true,
				PreConfig:          func() { e2etestutils.V1Apply(t, []manifest.Object{manifestProject}) },
			},
			// Update and Read, ensure computed field does not pollute the plan.
			{
				Config: newProjectResource(t, func() projectResourceTemplateModel {
					m := projectResource
					m.DisplayName = types.StringValue("New Project Display Name")
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_project.test", "display_name", "New Project Display Name"),
					assertResourceWasApplied(t, ctx, func() v1alphaProject.Project {
						project := manifestProject
						project.Metadata.DisplayName = "New Project Display Name"
						return project
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"display_name"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_project.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Update name and revert display name - recreate.
			{
				Config: newProjectResource(t, func() projectResourceTemplateModel {
					m := projectResource
					m.Name = projectNameRecreatedByNameChange
					return m
				}()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_project.test", "name", projectNameRecreatedByNameChange),
					assertResourceWasApplied(t, ctx, func() v1alphaProject.Project {
						project := manifestProject
						project.Metadata.Name = projectNameRecreatedByNameChange
						return project
					}()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						expectChangesInResourcePlan(planDiff{Modified: []string{"name", "display_name"}}),
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_project.test", plancheck.ResourceActionReplace),
					},
				},
			},
			// Delete automatically occurs in TestCase, no need to clean up.
		},
	})
}

func TestAccProjectResource_planValidation(t *testing.T) {
	t.Parallel()
	testAccSetup(t)

	projectResource := projectResourceTemplateModel{
		ResourceName:         "test",
		ProjectResourceModel: getExampleProjectResource(t),
	}
	projectResource.Name = "not valid"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      newProjectResource(t, projectResource),
				ExpectError: regexp.MustCompile(`Bad Request: Validation for Project 'not valid' has failed`),
				PlanOnly:    true,
			},
		},
	})
}

func TestRenderProjectResourceTemplate(t *testing.T) {
	t.Parallel()

	exampleResource := getExampleProjectResource(t)
	exampleResource.Name = "project"
	exampleResource.Labels = Labels{
		{Key: "team", Values: []string{"green", "orange"}},
		{Key: "env", Values: []string{"prod"}},
		{Key: "empty", Values: []string{""}},
	}
	actual := newProjectResource(t, projectResourceTemplateModel{
		ResourceName:         "this",
		ProjectResourceModel: exampleResource,
	})

	assertHCL(t, actual)
	assert.Equal(t, readExpectedConfig(t, "project-config.tf"), actual)
}

type projectResourceTemplateModel struct {
	ResourceName string
	ProjectResourceModel
}

func newProjectResource(t *testing.T, model projectResourceTemplateModel) string {
	return executeTemplate(t, "project_resource.hcl.tmpl", model)
}

func getExampleProjectResource(t *testing.T) ProjectResourceModel {
	return ProjectResourceModel{
		Name:        e2etestutils.GenerateName(),
		DisplayName: types.StringValue("Project"),
		Description: types.StringValue("Example project"),
		Annotations: map[string]string{"key": "value"},
		Labels: addTestLabels(t, Labels{
			{Key: "team", Values: []string{"green"}},
			{Key: "env", Values: []string{"dev", "prod"}},
		}),
	}
}

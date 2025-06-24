package frameworkprovider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/stretchr/testify/assert"
)

func TestAccProjectResource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	projectName := generateName()
	projectNameRecreatedByNameChange := generateName()

	projectResource := projectResourceTemplateModel{
		ResourceName:         "test",
		ProjectResourceModel: getExampleProjectResource(t),
	}
	projectResource.ProjectResourceModel.Labels = appendTestLabels(projectResource.ProjectResourceModel.Labels)
	projectResource.ProjectResourceModel.Name = projectName

	manifestProject := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name:        projectName,
			DisplayName: "Project",
			Annotations: v1alpha.MetadataAnnotations{"key": "value"},
			Labels: annotateV1alphaLabels(t, v1alpha.Labels{
				"team": []string{"green"},
				"env":  []string{"dev", "prod"},
			}),
		},
		v1alphaProject.Spec{
			Description: "Example project",
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
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
				ImportStateId: projectName,
				ImportState:   true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if !assert.Len(t, states, 1) {
						return errors.New("expected exactly one state")
					}
					assert.Equal(t, projectName, states[0].Attributes["name"])
					return nil
				},
				// In the next step we're also verifying the imported state, so we need to persist it.
				ImportStatePersist: true,
				PreConfig:          func() { applyNobl9Objects(t, ctx, manifestProject) },
			},
			// Update and Read, ensure computed field does not pollute the plan.
			{
				Config: newProjectResource(t, func() projectResourceTemplateModel {
					m := projectResource
					m.ProjectResourceModel.DisplayName = types.StringValue("New Project Display Name")
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
						expectNoChangeInPlan{attrName: "status"},
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_project.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Update name - recreate.
			{
				Config: newProjectResource(t, func() projectResourceTemplateModel {
					m := projectResource
					m.ProjectResourceModel.Name = projectNameRecreatedByNameChange
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
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_project.test", plancheck.ResourceActionReplace),
					},
				},
			},
			// Delete automatically occurs in TestCase, no need to clean up.
		},
	})
}

func TestRenderProjectResourceTemplate(t *testing.T) {
	t.Parallel()

	actual := newProjectResource(t, projectResourceTemplateModel{
		ResourceName:         "this",
		ProjectResourceModel: getExampleProjectResource(t),
	})

	expected := fmt.Sprintf(`resource "nobl9_project" "this" {
  name = "project"
  display_name = "Project"
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
  description = "Example project"
}
`, testStartTime.UnixNano(), t.Name())

	assert.Equal(t, expected, actual)
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
		Name:        "project",
		DisplayName: types.StringValue("Project"),
		Description: types.StringValue("Example project"),
		Annotations: map[string]string{"key": "value"},
		Labels: annotateLabels(t, Labels{
			{Key: "team", Values: []string{"green"}},
			{Key: "env", Values: []string{"prod", "dev"}},
		}),
	}
}

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
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/stretchr/testify/assert"
)

func TestAccProjectResource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	unixNow := time.Now().UnixNano()
	projectName := fmt.Sprintf("project-%d", unixNow)
	projectNameRecreatedByNameChange := fmt.Sprintf("project-name-recreated-%d", unixNow)

	projectResource := projectResourceTemplateModel{
		ResourceName:         "test",
		ProjectResourceModel: getExampleProjectResource(),
	}
	projectResource.ProjectResourceModel.Labels = appendTestLabels(projectResource.ProjectResourceModel.Labels)
	projectResource.ProjectResourceModel.Name = projectName

	manifestProject := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name:        projectName,
			DisplayName: "Project",
			Annotations: v1alpha.MetadataAnnotations{"key": "value"},
			Labels: v1alpha.Labels{
				"team":   []string{"green"},
				"env":    []string{"dev", "prod"},
				"origin": []string{"terraform-acc-test"},
			},
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
			// ImportState - invalid id.
			{
				ResourceName:  fmt.Sprintf("nobl9_project.test_", unixNow),
				ImportStateId: projectName,
				ImportState:   true,
				ExpectError:   regexp.MustCompile(`Invalid import ID`),
			},
			// ImportState.
			{
				ResourceName:  "nobl9_project.test",
				ImportStateId: "default/" + projectName,
				ImportState:   true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if !assert.Len(t, states, 1) {
						return errors.New("expected exactly one state")
					}
					assert.Equal(t, projectName, states[0].Attributes["name"])
					assert.Equal(t, "default", states[0].Attributes["project"])
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
			// Update project - recreate.
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
		ProjectResourceModel: getExampleProjectResource(),
	})

	expected := `resource "nobl9_project" "this" {
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
  description = "Example project"
}
`

	assert.Equal(t, expected, actual)
}

type projectResourceTemplateModel struct {
	ResourceName string
	ProjectResourceModel
}

func newProjectResource(t *testing.T, model projectResourceTemplateModel) string {
	return executeTemplate(t, "project_resource.hcl.tmpl", model)
}

func getExampleProjectResource() ProjectResourceModel {
	return ProjectResourceModel{
		Name:        "project",
		DisplayName: types.StringValue("Project"),
		Description: types.StringValue("Example project"),
		Annotations: map[string]string{"key": "value"},
		Labels: Labels{
			{Key: "team", Values: []string{"green"}},
			{Key: "env", Values: []string{"prod", "dev"}},
		},
	}
}

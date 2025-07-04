package frameworkprovider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

func TestAccProjectDataSource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	unixNow := time.Now().UnixNano()

	projectName := fmt.Sprintf("project-%d", unixNow)
	manifestProject := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name: projectName,
		},
		v1alphaProject.Spec{},
	)

	serviceName := fmt.Sprintf("service-%d", unixNow)
	manifestService := v1alphaService.New(
		v1alphaService.Metadata{
			Name:    serviceName,
			Project: projectName,
		},
		v1alphaService.Spec{},
	)
	manifestService.Status = &v1alphaService.Status{
		SloCount: 0,
	}

	serviceResourceConfig := executeTemplate(t, "project_data_source.hcl.tmpl", map[string]any{
		"DataSourceName": "test",
		"ResourceName":   "test",
		"ProjectName":    projectName,
		"ServiceName":    serviceName,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create Service resource with Project Data Source.
			{
				PreConfig: func() {
					applyNobl9Objects(t, ctx, manifestProject)
					t.Cleanup(func() {
						deleteNobl9Objects(t, ctx, manifestProject)
					})
				},
				Config: serviceResourceConfig,
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
			// Delete automatically occurs in TestCase, no need to clean up.
		},
	})
}

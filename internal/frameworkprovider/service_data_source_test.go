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

func TestAccServiceDataSource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	unixNow := time.Now().UnixNano()

	project1Name := fmt.Sprintf("project-1-%d", unixNow)
	manifestProject1 := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name: project1Name,
		},
		v1alphaProject.Spec{},
	)

	project2Name := fmt.Sprintf("project-2-%d", unixNow)
	manifestProject2 := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name: project2Name,
		},
		v1alphaProject.Spec{},
	)

	serviceName := fmt.Sprintf("service-%d", unixNow)
	manifestService1 := v1alphaService.New(
		v1alphaService.Metadata{
			Name:    serviceName,
			Project: project1Name,
		},
		v1alphaService.Spec{},
	)
	manifestService1.Status = &v1alphaService.Status{
		SloCount: 0,
	}

	manifestService2 := manifestService1
	manifestService2.Metadata.Project = project2Name

	serviceResourceConfig := executeTemplate(t, "service_data_source.hcl.tmpl", map[string]any{
		"Project1Name": project1Name,
		"Project2Name": project2Name,
		"ServiceName":  serviceName,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Create Service resource with Service Data Source name in another Project.
			{
				PreConfig: func() {
					applyNobl9Objects(t, ctx, manifestProject1, manifestProject2, manifestService1)
					t.Cleanup(func() {
						deleteNobl9Objects(t, ctx, manifestProject1, manifestProject2, manifestService1)
					})
				},
				Config: serviceResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasApplied(t, ctx, manifestService2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction("nobl9_service.this", plancheck.ResourceActionCreate),
					},
				},
			},
			// Delete automatically occurs in TestCase, no need to clean up.
		},
	})
}

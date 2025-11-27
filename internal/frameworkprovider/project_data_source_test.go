package frameworkprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func TestAccProjectDataSource(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name: e2etestutils.GenerateName(),
		},
		v1alphaProject.Spec{},
	)

	manifestService := v1alphaService.New(
		v1alphaService.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: manifestProject.GetName(),
		},
		v1alphaService.Spec{},
	)

	serviceResource := newServiceResourceConfigFromManifest(manifestService)
	serviceResourceConfig := executeTemplate(t, "project_data_source.hcl.tmpl", map[string]any{
		"Project": newProjectResourceConfigFromManifest(manifestProject),
		"Service": serviceResource,
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create Service resource with Project Data Source.
			{
				PreConfig: func() {
					e2etestutils.V1Apply(t, []manifest.Object{manifestProject})
					t.Cleanup(func() {
						e2etestutils.V1Delete(t, []manifest.Object{manifestProject})
					})
				},
				Config: serviceResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					assertResourceWasApplied(t, ctx, manifestService),
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

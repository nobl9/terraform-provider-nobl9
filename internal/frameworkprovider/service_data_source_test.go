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

func TestAccServiceDataSource(t *testing.T) {
	t.Parallel()
	testAccSetup(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	manifestProject1 := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name:   e2etestutils.GenerateName(),
			Labels: e2etestutils.AnnotateLabels(t, nil),
		},
		v1alphaProject.Spec{},
	)

	manifestProject2 := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name:   e2etestutils.GenerateName(),
			Labels: e2etestutils.AnnotateLabels(t, nil),
		},
		v1alphaProject.Spec{},
	)

	manifestService1 := v1alphaService.New(
		v1alphaService.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: manifestProject1.GetName(),
			Labels:  e2etestutils.AnnotateLabels(t, nil),
		},
		v1alphaService.Spec{},
	)
	manifestService1.Status = &v1alphaService.Status{
		SloCount: 0,
	}

	manifestService2 := manifestService1
	manifestService2.Metadata.Project = manifestProject2.GetName()

	serviceResourceConfig := executeTemplate(t, "service_data_source.hcl.tmpl", map[string]any{
		"Project1Name": manifestProject1.GetName(),
		"Project2Name": manifestProject2.GetName(),
		"ServiceName":  manifestService1.GetName(),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create Service resource with Service Data Source name in another Project.
			{
				PreConfig: func() {
					objects := []manifest.Object{manifestProject1, manifestProject2, manifestService1}
					e2etestutils.V1Apply(t, objects)
					t.Cleanup(func() {
						e2etestutils.V1Delete(t, objects)
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

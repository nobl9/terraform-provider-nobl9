package nobl9

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

func resourceComposite() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max_delay": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Maximum time for your composite SLO to wait for data from objectives.",
			},
			"components": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Objectives to be assembled in your composite SLO.",
				Elem:        resourceCompositeComponents(),
			},
		},
	}
}

func resourceCompositeComponents() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"objectives": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "An additional nesting for the components of your composite SLO.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"composite_objective": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Your composite SLO component.",
							Elem:        resourceCompositeObjective(),
						},
					},
				},
			},
		},
	}
}

func resourceCompositeObjective() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project name.",
			},
			"slo": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SLO name.",
			},
			"objective": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SLO objective name.",
			},
			"weight": {
				Type:        schema.TypeFloat,
				Required:    true,
				Description: "Weights determine each component’s contribution to the composite SLO.",
			},
			"when_delayed": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Defines how to treat missing component data on `max_delay` expiry.",
				ValidateFunc: validation.StringInSlice(v1alphaSLO.WhenDelayedNames(), false),
			},
		},
	}
}

func schemaCompositeDeprecated() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "(\"composite\" is deprecated, use [composites 2.0 schema](https://registry.terraform.io/providers/nobl9/nobl9/latest/docs/resources/slo#nested-schema-for-objectivecomposite) instead) [Composite SLO documentation](https://docs.nobl9.com/yaml-guide/#slo)",
		Deprecated:  "\"composite\" is deprecated, use \"objective.composite\" instead.",
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"target": {
					Type:        schema.TypeFloat,
					Required:    true,
					Description: "Designated value.",
				},
				"burn_rate_condition": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "(\"burn_rate_condition\" is part of deprecated composites 1.0, use [composites 2.0](https://registry.terraform.io/providers/nobl9/nobl9/latest/docs/resources/slo#nested-schema-for-objectivecomposite) instead) Condition when the Composite SLO’s error budget is burning.",
					Deprecated:  "\"burn_rate_condition\" is part of deprecated composites 1.0, use composites 2.0 (https://registry.terraform.io/providers/nobl9/nobl9/latest/docs/resources/slo#nested-schema-for-objectivecomposite) instead",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"op": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Type of logical operation.",
							},
							"value": {
								Type:        schema.TypeFloat,
								Required:    true,
								Description: "Burn rate value.",
							},
						},
					},
				},
			},
		},
	}
}

func marshalComposite(sloObjective map[string]interface{}) (*v1alphaSLO.CompositeSpec, error) {
	compositeSet := sloObjective["composite"].(*schema.Set)
	if compositeSet.Len() == 0 {
		return nil, nil
	}
	compositeMap := compositeSet.List()[0].(map[string]interface{})
	maxDelay := compositeMap["max_delay"].(string)

	componentsSet := compositeMap["components"].(*schema.Set)
	resultObjectives := make([]v1alphaSLO.CompositeObjective, 0)
	if len(componentsSet.List()) > 0 {
		componentsMap := componentsSet.List()[0].(map[string]interface{})
		objectivesSet := componentsMap["objectives"].(*schema.Set)
		compositeObjectivesMap := objectivesSet.List()[0].(map[string]interface{})
		compositeObjectives := compositeObjectivesMap["composite_objective"].([]interface{})

		for _, compObjElems := range compositeObjectives {
			compObj := compObjElems.(map[string]interface{})
			whenDelayed, ok := compObj["when_delayed"].(string)
			if !ok || whenDelayed == "" {
				return nil, fmt.Errorf("when_delayed is required for composite objective")
			}
			whenDelayedParsed, err := v1alphaSLO.ParseWhenDelayed(whenDelayed)
			if err != nil {
				return nil, err
			}

			resultObjectives = append(resultObjectives, v1alphaSLO.CompositeObjective{
				Project:     compObj["project"].(string),
				SLO:         compObj["slo"].(string),
				Objective:   compObj["objective"].(string),
				Weight:      compObj["weight"].(float64),
				WhenDelayed: whenDelayedParsed,
			})
		}
	}
	return &v1alphaSLO.CompositeSpec{
		MaxDelay:   maxDelay,
		Components: v1alphaSLO.Components{Objectives: resultObjectives},
	}, nil
}

func marshalCompositeDeprecated(d *schema.ResourceData) *v1alphaSLO.Composite {
	compositeSet := d.Get("composite").(*schema.Set)

	if compositeSet.Len() > 0 {
		compositeTf := compositeSet.List()[0].(map[string]interface{})

		var burnRateCondition *v1alphaSLO.CompositeBurnRateCondition
		burnRateConditionSet := compositeTf["burn_rate_condition"].(*schema.Set)

		if burnRateConditionSet.Len() > 0 {
			burnRateConditionTf := burnRateConditionSet.List()[0].(map[string]interface{})

			burnRateCondition = &v1alphaSLO.CompositeBurnRateCondition{
				Value:    burnRateConditionTf["value"].(float64),
				Operator: burnRateConditionTf["op"].(string),
			}
		}

		budgetTarget := compositeTf["target"].(float64)
		return &v1alphaSLO.Composite{
			BudgetTarget:      &budgetTarget,
			BurnRateCondition: burnRateCondition,
		}
	}

	return nil
}

func unmarshalComposite(compositeSpec *v1alphaSLO.CompositeSpec) *schema.Set {
	if compositeSpec == nil {
		return nil
	}

	composite := make(map[string]interface{})
	composite["max_delay"] = compositeSpec.MaxDelay

	compObjList := make([]interface{}, 0, len(compositeSpec.Components.Objectives))
	for _, objective := range compositeSpec.Components.Objectives {
		compositeObjective := make(map[string]interface{})
		compositeObjective["project"] = objective.Project
		compositeObjective["slo"] = objective.SLO
		compositeObjective["objective"] = objective.Objective
		compositeObjective["weight"] = objective.Weight
		compositeObjective["when_delayed"] = objective.WhenDelayed.String()

		compObjList = append(compObjList, compositeObjective)
	}

	objectives := make(map[string]interface{})
	objectives["composite_objective"] = compObjList

	components := make(map[string]interface{})
	components["objectives"] = schema.NewSet(
		schema.HashResource(resourceCompositeComponents()),
		[]interface{}{objectives},
	)
	composite["components"] = schema.NewSet(oneElementSet, []interface{}{components})

	return schema.NewSet(schema.HashResource(resourceComposite()), []interface{}{composite})
}

func unmarshalCompositeDeprecated(d *schema.ResourceData, spec v1alphaSLO.Spec) error {
	//nolint:staticcheck
	if spec.Composite != nil {
		//nolint:staticcheck
		composite := spec.Composite
		compositeTF := make(map[string]interface{})

		compositeTF["target"] = composite.BudgetTarget

		if composite.BurnRateCondition != nil {
			burnRateCondition := composite.BurnRateCondition
			burnRateConditionTF := make(map[string]interface{})
			burnRateConditionTF["value"] = burnRateCondition.Value
			burnRateConditionTF["op"] = burnRateCondition.Operator
			compositeTF["burn_rate_condition"] = schema.NewSet(oneElementSet, []interface{}{burnRateConditionTF})
		}

		return d.Set("composite", schema.NewSet(oneElementSet, []interface{}{compositeTF}))
	}

	return nil
}

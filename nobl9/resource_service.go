package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func ResourceService() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"apiVersion": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "API version",
			},

			"kind": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Kind of object",
			},

			"manifestSrc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
			},

			"metadata": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"displayName": {
							Type:        schema.TypeString,
							Description: "",
						},

						"labels": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},

						"name": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "",
						},

						"project": {
							Type:        schema.TypeString,
							Description: "",
						},
					},
				},
			},

			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
			},

			"spec": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:        schema.TypeString,
							Description: "",
						},
					},
				},
			},

			"status": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sloCount": {
							Type:        schema.TypeInt,
							Description: "",
						},
					},
				},
			},
		},
		CreateContext: CreateService,
		//UpdateContext: UpdateService,
		//DeleteContext: DeleteService,
		//ReadContext:   ReadService,
		Description: "* [HTTP API](https://api-docs.app.nobl9.com/)",
	}
}

func marshalService(d *schema.ResourceData) *n9api.Service {
	return &n9api.Service{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion: d.Get("apiVersion").(string),
			Kind:       d.Get("kind").(string),
			MetadataHolder: n9api.MetadataHolder{
				Metadata: n9api.Metadata{
					Name:        d.Get("name").(string),
					DisplayName: d.Get("displayName").(string),
					Project:     d.Get("proejct").(string),
				},
			},
		},
		Spec: n9api.ServiceSpec{
			Description: d.Get("description").(string),
		},
		Status: n9api.ServiceStatus{
			SloCount: d.Get("sloCount").(int),
		},
	}
}

func CreateService(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*n9api.Client)

	service := marshalService(d)
	var p n9api.Payload
	p.AddObject(service)

	err := c.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add service: %v", err)
	}

	//d.SetId(strconv.FormatInt(id, 10))

	return nil

}

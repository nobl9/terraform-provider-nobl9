package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceService() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"api_version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "API version",
			},

			"kind": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Kind of object",
			},

			"manifest_src": {
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
						"display_name": {
							Type:        schema.TypeString,
							Optional:    true,
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
							Optional:    true,
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
							Optional:    true,
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
						"slo_count": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "",
						},
					},
				},
			},
		},
		CreateContext: resourceServiceCreate,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		ReadContext:   resourceServiceRead,
		Description:   "* [HTTP API](https://api-docs.app.nobl9.com/)",
	}
}

func marshalService(d *schema.ResourceData) *n9api.Service {
	return &n9api.Service{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion: d.Get("api_version").(string),
			Kind:       d.Get("kind").(string),
			MetadataHolder: n9api.MetadataHolder{
				Metadata: n9api.Metadata{
					Name: d.Get("name").(string),
				},
			},
		},
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*n9api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	service := marshalService(d)
	var p n9api.Payload
	p.AddObject(service)

	err := c.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add service: %v", err)
	}

	//d.SetId(strconv.FormatInt(id, 10))

	return diags
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//c := meta.(*n9api.Client)

	//d.SetId(strconv.FormatInt(id, 10))

	return nil
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	//d.SetId(strconv.FormatInt(id, 10))

	return nil
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	//d.SetId(strconv.FormatInt(id, 10))

	return nil
}

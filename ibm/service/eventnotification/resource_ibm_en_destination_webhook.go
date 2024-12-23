// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package eventnotification

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	en "github.com/IBM/event-notifications-go-admin-sdk/eventnotificationsv1"
)

func ResourceIBMEnWebhookDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMEnWebhookDestinationCreate,
		ReadContext:   resourceIBMEnWebhookDestinationRead,
		UpdateContext: resourceIBMEnWebhookDestinationUpdate,
		DeleteContext: resourceIBMEnWebhookDestinationDelete,
		Importer:      &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"instance_guid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for IBM Cloud Event Notifications instance.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Destintion name.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of Destination Webhook.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Destination description.",
			},
			"collect_failed_events": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to collect the failed event in Cloud Object Storage bucket",
			},
			"config": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Payload describing a destination configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"params": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "URL of webhook.",
									},
									"verb": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "HTTP method of webhook.",
									},
									"custom_headers": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Custom headers (Key-Value pair) for webhook call.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									"sensitive_headers": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "List of sensitive headers from custom headers.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"destination_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Destination ID",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last updated time.",
			},
			"subscription_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of subscriptions.",
			},
			"subscription_names": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of subscriptions.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceIBMEnWebhookDestinationCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	enClient, err := meta.(conns.ClientSession).EventNotificationsApiV1()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, err.Error(), "ibm_en_destination_webhook", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
		// return diag.FromErr(err)
	}

	options := &en.CreateDestinationOptions{}

	options.SetInstanceID(d.Get("instance_guid").(string))
	options.SetName(d.Get("name").(string))
	options.SetType(d.Get("type").(string))
	options.SetCollectFailedEvents(d.Get("collect_failed_events").(bool))

	destinationtype := d.Get("type").(string)
	if _, ok := d.GetOk("description"); ok {
		options.SetDescription(d.Get("description").(string))
	}
	if _, ok := d.GetOk("config"); ok {
		config := WebhookdestinationConfigMapToDestinationConfig(d.Get("config.0.params.0").(map[string]interface{}), destinationtype)
		options.SetConfig(&config)
	}

	result, _, err := enClient.CreateDestinationWithContext(context, options)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("CreateDestinationWithContext failed: %s", err.Error()), "ibm_en_destination_webhook", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
		// return diag.FromErr(fmt.Errorf("CreateDestinationWithContext failed %s\n%s", err, response))
	}

	d.SetId(fmt.Sprintf("%s/%s", *options.InstanceID, *result.ID))

	return resourceIBMEnWebhookDestinationRead(context, d, meta)
}

func resourceIBMEnWebhookDestinationRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	enClient, err := meta.(conns.ClientSession).EventNotificationsApiV1()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, err.Error(), "ibm_en_destination_webhook", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
		// return diag.FromErr(err)
	}

	options := &en.GetDestinationOptions{}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		tfErr := flex.TerraformErrorf(err, err.Error(), "ibm_en_destination_webhook", "read")
		return tfErr.GetDiag()
		// return diag.FromErr(err)
	}

	options.SetInstanceID(parts[0])
	options.SetID(parts[1])

	result, response, err := enClient.GetDestinationWithContext(context, options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetDestinationWithContext failed: %s", err.Error()), "ibm_en_destination_webhook", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
		// return diag.FromErr(fmt.Errorf("GetDestinationWithContext failed %s\n%s", err, response))
	}

	if err = d.Set("instance_guid", options.InstanceID); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting instance_guid: %s", err))
	}

	if err = d.Set("destination_id", options.ID); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting destination_id: %s", err))
	}

	if err = d.Set("name", result.Name); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting name: %s", err))
	}

	if err = d.Set("type", result.Type); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting type: %s", err))
	}

	if err = d.Set("description", result.Description); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting description: %s", err))
	}

	if result.CollectFailedEvents != nil {
		if err = d.Set("collect_failed_events", result.CollectFailedEvents); err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error setting CollectFailedEvents: %s", err))
		}
	}

	if result.Config != nil {
		err = d.Set("config", enWebhookDestinationFlattenConfig(*result.Config))
		if err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error setting config %s", err))
		}
	}

	if err = d.Set("updated_at", flex.DateTimeToString(result.UpdatedAt)); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting updated_at: %s", err))
	}

	if err = d.Set("subscription_count", flex.IntValue(result.SubscriptionCount)); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting subscription_count: %s", err))
	}

	if err = d.Set("subscription_names", result.SubscriptionNames); err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error setting subscription_names: %s", err))

	}

	return nil
}

func resourceIBMEnWebhookDestinationUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	enClient, err := meta.(conns.ClientSession).EventNotificationsApiV1()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, err.Error(), "ibm_en_destination_webhook", "update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
		// return diag.FromErr(err)
	}

	options := &en.UpdateDestinationOptions{}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		tfErr := flex.TerraformErrorf(err, err.Error(), "ibm_en_destination_webhook", "update")
		return tfErr.GetDiag()
		// return diag.FromErr(err)
	}

	options.SetInstanceID(parts[0])
	options.SetID(parts[1])

	if ok := d.HasChanges("name", "description", "collect_failed_events", "config"); ok {
		options.SetName(d.Get("name").(string))

		if _, ok := d.GetOk("description"); ok {
			options.SetDescription(d.Get("description").(string))
		}

		if _, ok := d.GetOk("collect_failed_events"); ok {
			options.SetCollectFailedEvents(d.Get("collect_failed_events").(bool))
		}
		destinationtype := d.Get("type").(string)
		if _, ok := d.GetOk("config"); ok {
			config := WebhookdestinationConfigMapToDestinationConfig(d.Get("config.0.params.0").(map[string]interface{}), destinationtype)
			options.SetConfig(&config)
		}
		_, _, err := enClient.UpdateDestinationWithContext(context, options)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("UpdateDestinationWithContext failed: %s", err.Error()), "ibm_en_destination_webhook", "update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
			// return diag.FromErr(fmt.Errorf("UpdateDestinationWithContext failed %s\n%s", err, response))
		}

		return resourceIBMEnWebhookDestinationRead(context, d, meta)
	}

	return nil
}

func resourceIBMEnWebhookDestinationDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	enClient, err := meta.(conns.ClientSession).EventNotificationsApiV1()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, err.Error(), "ibm_en_destination_webhook", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
		// return diag.FromErr(err)
	}

	options := &en.DeleteDestinationOptions{}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		tfErr := flex.TerraformErrorf(err, err.Error(), "ibm_en_destination_webhook", "delete")
		return tfErr.GetDiag()
		// return diag.FromErr(err)
	}

	options.SetInstanceID(parts[0])
	options.SetID(parts[1])

	response, err := enClient.DeleteDestinationWithContext(context, options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("DeleteDestinationWithContext failed: %s", err.Error()), "ibm_en_destination_webhook", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
		// return diag.FromErr(fmt.Errorf("DeleteDestinationWithContext failed %s\n%s", err, response))
	}

	d.SetId("")

	return nil
}

func WebhookdestinationConfigMapToDestinationConfig(configParams map[string]interface{}, destinationtype string) en.DestinationConfig {
	params := new(en.DestinationConfigOneOf)
	if configParams["url"] != nil {
		params.URL = core.StringPtr(configParams["url"].(string))
	}

	if configParams["verb"] != nil {
		params.Verb = core.StringPtr(configParams["verb"].(string))
	}

	if configParams["custom_headers"] != nil {
		var customHeaders = make(map[string]string)
		for k, v := range configParams["custom_headers"].(map[string]interface{}) {
			customHeaders[k] = v.(string)
		}
		params.CustomHeaders = customHeaders
	}

	if configParams["sensitive_headers"] != nil {
		sensitiveHeaders := []string{}
		for _, sensitiveHeadersItem := range configParams["sensitive_headers"].([]interface{}) {
			sensitiveHeaders = append(sensitiveHeaders, sensitiveHeadersItem.(string))
		}
		params.SensitiveHeaders = sensitiveHeaders
	}
	destinationConfig := new(en.DestinationConfig)
	destinationConfig.Params = params
	return *destinationConfig
}

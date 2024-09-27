// This output allows config_aggregator_settings data to be referenced by other resources and the terraform CLI
// Modify this output if only certain data should be exposed
output "config_aggregator_settings" {
  value = {
    additional_scope            = []
    regions                     = ["all"]
    resource_collection_enabled  = ibm_config_aggregator_settings.config_aggregator_settings_instance.resource_collection_enabled
    trusted_profile_id           = ibm_config_aggregator_settings.config_aggregator_settings_instance.trusted_profile_id
  }
}

output "aggregator_settings" {
  value = {
    additional_scope            = []
    regions                     = ["all"]
    resource_collection_enabled  = ibm_config_aggregator_settings.config_aggregator_settings_instance.resource_collection_enabled
    trusted_profile_id           = ibm_config_aggregator_settings.config_aggregator_settings_instance.trusted_profile_id
  }
}

output "config_aggregator_configurations" {
  value = {
    about = {
      account_id               = data.ibm_config_aggregator_configurations.config_aggregator_configurations_instance.account_id
      config_type              = data.ibm_config_aggregator_configurations.config_aggregator_configurations_instance.config_type
      location                 = data.ibm_config_aggregator_configurations.config_aggregator_configurations_instance.location
      resource_crn             = data.ibm_config_aggregator_configurations.config_aggregator_configurations_instance.resource_crn
      resource_group_id        = data.ibm_config_aggregator_configurations.config_aggregator_configurations_instance.resource_group_id
      resource_name            = data.ibm_config_aggregator_configurations.config_aggregator_configurations_instance.resource_name
      service_name             = data.ibm_config_aggregator_configurations.config_aggregator_configurations_instance.service_name
    },
    config = {
      event_notification_enabled = data.ibm_config_aggregator_configurations.config_aggregator_configurations_instance.event_notification_enabled
    }
  }
}

output "config_aggregator_resource_collection_status"{
    value={
        status = "complete"
    }
}
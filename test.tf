terraform {
  required_version = ">=1.3"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.106.1" # 3.40.0 doesn't work
    }
    azapi = {
      source = "azure/azapi"
    }
  }
}

provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }
}

provider "azapi" {
}

resource "random_string" "random" {
  length  = 5
  upper   = false
  special = false
}

resource "azurerm_resource_group" "group" {
  name = "test-${random_string.random.result}"
  location = "polandcentral"
}

module "aks" {
  source = "Azure/aks/azurerm"
  version = "7.5.0"

  kubernetes_version   = "1.22"
  cluster_name         = "marcin"
  resource_group_name  = azurerm_resource_group.group.name
  prefix               = "marcin"
  os_disk_size_gb      = 60
  sku_tier             = "Standard"
  rbac_aad             = false
  # vnet_subnet_id       = azurerm_subnet.network.id
  # node_pools           = {for name, pool in var.node_pools : name => merge(pool, {name = name, vnet_subnet_id = azurerm_subnet.network.id})}

  ebpf_data_plane     = "cilium"
  network_plugin_mode = "overlay"
  network_plugin      = "azure"

  role_based_access_control_enabled = true

  workload_identity_enabled = true
  oidc_issuer_enabled       = true
}

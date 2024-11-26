terraform {
  required_version = ">=1.3"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">=3.51.0, < 4.0"
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

resource "random_string" "random" {
  length  = 5
  upper   = false
  special = false
}

data "azurerm_resource_group" "group" {
  name = "test-${random_string.random.result}"
}

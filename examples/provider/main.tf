terraform {
  required_providers {
    registry = {
      source  = "registry.terraform.io/terraform-registry/registry"
      version = "~> 0.1"
    }
  }
}

provider "registry" {
  endpoint = "https://registry.example.com"
  token    = var.registry_token
}

variable "registry_token" {
  description = "API key or JWT token for the registry."
  type        = string
  sensitive   = true
}

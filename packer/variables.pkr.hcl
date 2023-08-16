variable "aws_region" {
  type    = string
  default = "us-east-2"
}

variable "img_name" {
  type    = string
  default = "plural/ubuntu-22.04"
}

variable "cli_version" {
  type = string
}

locals {
  cli_version_clean = trim(var.cli_version, "v")
}

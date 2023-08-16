variable "aws_target_regions" {
  type    = list(string)
  default = ["us-east-1", "us-east-2", "us-west-2", "ap-southeast-2"]
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

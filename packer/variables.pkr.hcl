variable "aws_target_regions" {
  type    = list(string)
  default = [
    "us-east-1",
    "us-east-2",
    "us-west-1",
    "us-west-2",
    "ca-central-1",
    "eu-central-1",
    "eu-west-1",
    "eu-west-2",
    "eu-west-3",
    "eu-north-1",
    "ap-northeast-1",
    "ap-northeast-2",
    "ap-northeast-3",
    "ap-south-1",
    "ap-southeast-1",
    "ap-southeast-2",
    "sa-east-1"
  ]
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

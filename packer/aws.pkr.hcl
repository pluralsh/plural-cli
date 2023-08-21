source "amazon-ebs" "main" {
  ami_name      = "${var.img_name}/${var.cli_version}"
  instance_type = "t2.micro"
  region        = "us-east-1"
  ami_regions   = var.aws_target_regions
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  run_tags = {
    Creator = "Packer"
  }
  run_volume_tags = {
    Creator = "Packer"
  }
  snapshot_tags = {
    Creator = "Packer"
  }
  tags = {
    Creator = "Packer"
  }
  ssh_username = "ubuntu"
  ami_groups = ["all"]
  force_deregister = true
  force_delete_snapshot = true
}

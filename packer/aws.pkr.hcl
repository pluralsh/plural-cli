source "amazon-ebs" "us-east-1" {
  ami_name      = var.img_name
  instance_type = "t2.micro"
  region        = "us-east-1"
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  ssh_username = "ubuntu"
  ami_groups = ["all"]
  force_deregister = true
  force_delete_snapshot = true
}

source "amazon-ebs" "us-east-2" {
  ami_name      = var.img_name
  instance_type = "t2.micro"
  region        = "us-east-2"
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  ssh_username = "ubuntu"
  ami_groups = ["all"]
  force_deregister = true
  force_delete_snapshot = true
}

source "amazon-ebs" "us-west-2" {
  ami_name      = var.img_name
  instance_type = "t2.micro"
  region        = "us-west-2"
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  ssh_username = "ubuntu"
  ami_groups = ["all"]
  force_deregister = true
  force_delete_snapshot = true
}

source "amazon-ebs" "ap-southeast-2" {
  ami_name      = var.img_name
  instance_type = "t2.micro"
  region        = "ap-southeast-2"
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  ssh_username = "ubuntu"
  ami_groups = ["all"]
  force_deregister = true
  force_delete_snapshot = true
}
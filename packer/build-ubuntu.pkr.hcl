build {
  name = "install-plural"
  sources = [
    "source.amazon-ebs.us-east-1",
    "source.amazon-ebs.us-east-2",
    "source.amazon-ebs.us-west-2",
    "source.amazon-ebs.ap-southeast-2",
  ]

  provisioner "shell" {
    inline = [
      "curl -L https://github.com/pluralsh/plural-cli/releases/download/${var.cli_version}/plural-cli_${local.cli_version_clean}_Linux_amd64.tar.gz | tar xvz plural",
      "chmod +x plural",
      "sudo mv plural /usr/local/bin/plural",
      "plural --help",
    ]
  }

  provisioner "shell" {
    inline = [
      "sudo apt-get update && sudo apt-get install -y gnupg software-properties-common curl unzip",
      "curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -",
      "sudo apt-add-repository \"deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main\"",
      "sudo apt-get update && sudo apt-get install terraform",
      "terraform --help",
    ]
  }

  provisioner "shell" {
    inline = [
      "curl https://baltocdn.com/helm/signing.asc | sudo apt-key add -",
      "sudo apt-get install apt-transport-https --yes",
      "echo \"deb https://baltocdn.com/helm/stable/debian/ all main\" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list",
      "sudo apt-get update",
      "sudo apt-get install helm",
      "helm --help",
      "helm plugin install https://github.com/pluralsh/helm-push",
      "helm plugin install https://github.com/databus23/helm-diff",
      "helm cm-push --help",
    ]
  }

  provisioner "shell" {
    inline = [
      "sudo snap install kubectl --classic",
      "kubectl version --client",
    ]
  }

  provisioner "shell" {
    inline = [
      "curl -L -o 'awscliv2.zip' 'https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip'",
      "unzip awscliv2.zip",
      "sudo ./aws/install",
      "aws --version",
    ]
  }

  provisioner "shell" {
    inline = [
      "sudo snap install google-cloud-sdk --classic",
      "gcloud --help",
    ]
  }

  provisioner "shell" {
    inline = [
      "curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash",
      "az --help",
    ]
  }

  post-processor "manifest" {
    output     = "manifest.json"
    strip_path = true
    custom_data = {
      image_name = "${var.img_name}/${var.cli_version}"
    }
  }
}

packer {
  required_version = ">= 1.7.0, < 2.0.0"

  required_plugins {
    qemu = {
      source  = "github.com/hashicorp/qemu"
      version = ">= 1.0.0, < 2.0.0"
    }
    sshkey = {
      version = ">= 1.0.1"
      source = "github.com/ivoronin/sshkey"
    }
  }
}

variable "cli_version" {
  type = string
  default = "v0.7.0"
}

locals {
  cli_version_clean = trim(var.cli_version, "v")
}

variable "boot_wait" {
  type    = string
  default = "6s"
}

variable "communicator" {
  type    = string
  default = "ssh"
}

variable "country" {
  type    = string
  default = "CA"
}

variable "cpus" {
  type    = string
  default = "1"
}

variable "disk_size" {
  type    = string
  default = "7500"
}

variable "headless" {
  type    = bool
  default = false
}

variable "host_port_max" {
  type    = string
  default = "4444"
}

variable "host_port_min" {
  type    = string
  default = "2222"
}

variable "http_port_max" {
  type    = string
  default = "9000"
}

variable "http_port_min" {
  type    = string
  default = "8000"
}

variable "iso_checksum" {
  type    = string
  default = "sha256:e099488c0d37a800be12e5df86a37ad29b74942330130374e71c8d734a20ca32"
}

variable "iso_file" {
  type    = string
  default = "jammy-server-cloudimg-amd64.img"
}

variable "iso_path_external" {
  type    = string
  default = "https://cloud-images.ubuntu.com/jammy/current"
}

variable "memory" {
  type    = string
  default = "1024"
}

variable "packer_cache_dir" {
  type    = string
  default = "${env("PACKER_CACHE_DIR")}"
}

variable "qemu_binary" {
  type    = string
  default = "qemu-system-x86_64"
#   default = "qemu-system-aarch64" # arm64
}

variable "shutdown_timeout" {
  type    = string
  default = "10m"
}

variable "ssh_agent_auth" {
  type    = bool
  default = false
}

variable "ssh_clear_authorized_keys" {
  type    = bool
  default = true
}

variable "ssh_disable_agent_forwarding" {
  type    = bool
  default = false
}

variable "ssh_file_transfer_method" {
  type    = string
  default = "scp"
}

variable "ssh_handshake_attempts" {
  type    = string
  default = "100"
}

variable "ssh_keep_alive_interval" {
  type    = string
  default = "5s"
}

data "sshkey" "install" {
}

variable "ssh_port" {
  type    = string
  default = "22"
}

variable "ssh_pty" {
  type    = bool
  default = true
}

variable "ssh_timeout" {
  type    = string
  default = "60m"
}

variable "ssh_username" {
  type    = string
  default = "ubuntu"
}

variable "start_retry_timeout" {
  type    = string
  default = "5m"
}

variable "vm_name" {
  type    = string
  default = "base-uefi-jammy"
}

variable "vnc_vrdp_bind_address" {
  type    = string
  default = "127.0.0.1"
}

variable "vnc_vrdp_port_max" {
  type    = string
  default = "6000"
}

variable "vnc_vrdp_port_min" {
  type    = string
  default = "5900"
}

locals {
  output_directory = "build"
}

source "qemu" "qemu" {
  boot_command         = []
  boot_wait            = var.boot_wait
  communicator         = var.communicator
  cpus                 = var.cpus
  disk_cache           = "writeback"
  disk_compression     = false
  disk_discard         = "ignore"
  disk_image           = true
  display              = "none"
  disk_interface       = "virtio-scsi"
  disk_size            = var.disk_size
  format               = "qcow2"
  headless             = var.headless
  host_port_max        = var.host_port_max
  host_port_min        = var.host_port_min
  http_port_max        = var.http_port_max
  http_port_min        = var.http_port_min
  iso_checksum         = var.iso_checksum
  iso_skip_cache       = false
  iso_target_path      = "${regex_replace(var.packer_cache_dir, "^$", "/tmp")}/${var.iso_file}"
  iso_urls = [
    "${var.iso_path_external}/${var.iso_file}"
  ]
  machine_type     = "q35"
  memory           = var.memory
  net_device       = "virtio-net"
  output_directory = local.output_directory
  qemu_binary      = var.qemu_binary
  cd_files = [
    "meta-data"
  ]
  cd_content = {
    "user-data" = templatefile("user-data", { ssh_public_key = data.sshkey.install.public_key })
  }
  cd_label = "cidata"
  shutdown_command             = "sudo shutdown -h now"
  shutdown_timeout             = var.shutdown_timeout
  skip_compaction              = true
  skip_nat_mapping             = false
  ssh_agent_auth               = var.ssh_agent_auth
  ssh_clear_authorized_keys    = var.ssh_clear_authorized_keys
  ssh_disable_agent_forwarding = var.ssh_disable_agent_forwarding
  ssh_file_transfer_method     = var.ssh_file_transfer_method
  ssh_handshake_attempts       = var.ssh_handshake_attempts
  ssh_keep_alive_interval      = var.ssh_keep_alive_interval
  ssh_private_key_file         = data.sshkey.install.private_key_path
  ssh_port                     = var.ssh_port
  ssh_pty                      = var.ssh_pty
  ssh_timeout                  = var.ssh_timeout
  ssh_username                 = var.ssh_username
  use_default_display          = true
  vm_name                      = var.vm_name
  vnc_bind_address             = var.vnc_vrdp_bind_address
  vnc_port_max                 = var.vnc_vrdp_port_max
  vnc_port_min                 = var.vnc_vrdp_port_min
}

build {
  description = "Ubuntu based VM image for KubeVirt"

  sources = ["source.qemu.qemu"]

  provisioner "shell" {
    inline = [

      # setup qemu-guest-agent
      "sudo apt-get update -yq",
      "sudo apt-get install -yq qemu-guest-agent ca-certificates curl gnupg",
      # "sudo systemctl start qemu-guest-agent",
      "sudo systemctl enable qemu-guest-agent",

      # install docker
      "sudo install -m 0755 -d /etc/apt/keyrings",
      "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg",
      "sudo chmod a+r /etc/apt/keyrings/docker.gpg",
      "echo \"deb [arch=\"$(dpkg --print-architecture)\" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \"$(. /etc/os-release && echo \"$VERSION_CODENAME\")\" stable\" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null",
      "sudo apt-get update",
      "VERSION_STRING=5:24.0.0-1~ubuntu.22.04~jammy",
      "sudo apt-get install -yq docker-ce=$VERSION_STRING docker-ce-cli=$VERSION_STRING containerd.io docker-buildx-plugin docker-compose-plugin systemd-container",



      # add plural group and user
      "sudo groupadd -g 1001 plural",
      "sudo useradd -ms /bin/bash -u 1001 -g plural plural",
      "sudo loginctl enable-linger plural",

      # setup for rootless docker
      "sudo apt-get -yq install uidmap dbus-user-session",
      "sudo machinectl shell plural@ /usr/bin/dockerd-rootless-setuptool.sh install",
      "sudo machinectl shell plural@ /usr/bin/systemctl --user start docker",
      "sudo machinectl shell plural@ /usr/bin/systemctl --user enable docker",

    ]
    inline_shebang = "/usr/bin/env bash"
    start_retry_timeout = var.start_retry_timeout
  }

  provisioner "shell" {
    inline = [
      "curl -L https://github.com/pluralsh/plural-cli/releases/download/${var.cli_version}/plural-cli_console_${local.cli_version_clean}_Linux_amd64.tar.gz | tar xvz plural",
      "chmod +x plural",
      "sudo mv plural /usr/local/bin/plural",
    ]
    inline_shebang = "/usr/bin/env bash"
    start_retry_timeout = var.start_retry_timeout
  }

  post-processors {
    post-processor "shell-local" {
      inline = [
        "docker build --platform linux/amd64 --push -f \"Dockerfile\" -t davidspek/plural-cli-kubevirt:0.1.13 .",
      ]
    }
  }
}

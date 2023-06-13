// If there is not specifying vpc_id, the module will launch a new vpc
resource "alicloud_vpc" "vpc" {
  count      = var.vpc_id == "" ? 1 : 0
  cidr_block = var.vpc_cidr
}

// According to the vswitch cidr blocks to launch several vswitches
resource "alicloud_vswitch" "vswitches" {
  count      = length(var.vswitch_ids) > 0 ? 0 : length(var.vswitch_cidrs)
  vpc_id     = var.vpc_id == "" ? join("", alicloud_vpc.vpc.*.id) : var.vpc_id
  cidr_block = element(var.vswitch_cidrs, count.index)
  zone_id    = element(var.availability_zone, count.index)
}

resource "alicloud_cs_kubernetes" "k8s" {
  count                 = 1
  master_vswitch_ids    = length(var.vswitch_ids) > 0 ? split(",", join(",", var.vswitch_ids)) : length(var.vswitch_cidrs) < 1 ? [] : split(",", join(",", alicloud_vswitch.vswitches.*.id))
  master_instance_types = var.master_instance_types
  node_cidr_mask        = var.node_cidr_mask
  enable_ssh            = var.enable_ssh
  install_cloud_monitor = var.install_cloud_monitor
  proxy_mode            = var.proxy_mode
  password              = var.password
  pod_cidr              = var.pod_cidr
  service_cidr          = var.service_cidr
  # version can not be defined in variables.tf.
  version = "1.24.6-aliyun.1"
  dynamic "addons" {
    for_each = var.cluster_addons
    content {
      name     = lookup(addons.value, "name", var.cluster_addons)
      config   = lookup(addons.value, "config", var.cluster_addons)
      disabled = lookup(addons.value, "disabled", var.cluster_addons)
    }
  }

}


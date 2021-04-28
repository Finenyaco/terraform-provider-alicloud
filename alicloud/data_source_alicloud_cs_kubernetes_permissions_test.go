package alicloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAlicloudCSKubernetesPermissionDataSource(t *testing.T) {
	rand := acctest.RandIntRange(1000000, 9999999)
	name := fmt.Sprintf("tf-testAccCSKubernetesPermission-%d", rand)

	resourceId := "data.alicloud_cs_kubernetes_permissions.default"
	testAccCheck := resourceAttrInit(resourceId, map[string]string{}).resourceAttrMapUpdateSet()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: dataSourceCSPermissionsConfigDependence(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"uid": CHECKSET,
					}),
				),
			},
		},
	})
}

func dataSourceCSPermissionsConfigDependence(name string) string {
	return fmt.Sprintf(`
variable "name" {
	default = "%s"
}

data "alicloud_zones" default {
  available_resource_creation  = "VSwitch"
}

data "alicloud_instance_types" "default" {
	availability_zone          = "${data.alicloud_zones.default.zones.0.id}"
	cpu_core_count             = 2
	memory_size                = 4
	kubernetes_node_role       = "Worker"
}

resource "alicloud_vpc" "default" {
  vpc_name                     = "${var.name}"
  cidr_block                   = "10.1.0.0/21"
}

resource "alicloud_vswitch" "default" {
  vswitch_name                 = "${var.name}"
  vpc_id                       = "${alicloud_vpc.default.id}"
  cidr_block                   = "10.1.1.0/24"
  availability_zone            = "${data.alicloud_zones.default.zones.0.id}"
}

# Create a management cluster
resource "alicloud_cs_managed_kubernetes" "default" {
  name                         = "${var.name}"
  count                        = 1
  cluster_spec                 = "ack.pro.small"
  is_enterprise_security_group = true
  worker_number                = 2
  deletion_protection          = false
  password                     = "Hello1234"
  pod_cidr                     = "172.20.0.0/16"
  service_cidr                 = "172.21.0.0/20"
  worker_vswitch_ids           = ["${alicloud_vswitch.default.id}"]
  worker_instance_types        = ["${data.alicloud_instance_types.default.instance_types.0.id}"]
  depends_on                   = ["alicloud_ram_user_policy_attachment.attach"]
}

# Create a new RAM user.
resource "alicloud_ram_user" "user" {
  name         = var.name
  display_name = var.name
  mobile       = "86-18688888888"
  email        = "hello.uuu@aaa.com"
  comments     = "yoyoyo"
}

# Create a new RAM Policy, .
resource "alicloud_ram_policy" "policy" {
  policy_name     = var.name
  policy_document = <<EOF
  {
    "Statement":[
      {
        "Action":"cs:Get*",
        "Effect":"Allow",
        "Resource":[
            "*"
        ]
      }
    ],
    "Version":"1"
  }
  EOF
  description = "this is a policy test by tf"
}

# Authorize the RAM user
resource "alicloud_ram_user_policy_attachment" "attach" {
  policy_name = alicloud_ram_policy.policy.name
  policy_type = alicloud_ram_policy.policy.type
  user_name   = alicloud_ram_user.user.name
}

# RBAC authorization for the cluster
resource "alicloud_cs_kubernetes_permissions" "default" {
  uid = alicloud_ram_user.user.id

  permissions {
    cluster     = alicloud_cs_managed_kubernetes.default.0.id
    role_type   = "cluster"
    role_name   = "dev"
    is_custom   = false
    is_ram_role = false
    namespace   = ""
  }
}
# Describe user permissions
data "alicloud_cs_kubernetes_permissions" "default" {
  uid = alicloud_cs_kubernetes_permissions.default.id
}
`, name)
}

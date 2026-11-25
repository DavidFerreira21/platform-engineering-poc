terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = ">= 5.83, < 6.0" }
  }
}
variable "name" { type = string }
variable "azs" {
  type = list(string)
  validation {
    condition     = length(var.azs) == 3 && length(distinct(var.azs)) == 3
    error_message = "Informe três AZs distintas em ordem estável."
  }
}
variable "config" {
  type = object({
    vpc       = object({ enabled = bool, cidr = string, dnsHostnames = bool, dnsSupport = bool })
    nat       = object({ enabled = bool })
    endpoints = object({ s3 = object({ enabled = bool }), ecr = object({ enabled = bool }) })
  })
}
locals {
  groups = ["web-nonpci", "web-pci", "database-nonpci", "database-pci"]
  subnets = merge([for group_index, group in local.groups : {
    for az_index, az in var.azs : "${group}-${az}" => {
      group = group
      az    = az
      cidr  = cidrsubnet(var.config.vpc.cidr, 4, group_index * 3 + az_index)
    }
  }]...)
}
resource "aws_vpc" "this" {
  cidr_block           = var.config.vpc.cidr
  enable_dns_hostnames = var.config.vpc.dnsHostnames
  enable_dns_support   = var.config.vpc.dnsSupport
  tags                 = { Name = var.name }
  lifecycle {
    precondition {
      condition     = var.config.vpc.enabled && try(tonumber(split("/", var.config.vpc.cidr)[1]) <= 24, false)
      error_message = "A stack de rede requer VPC habilitada e CIDR com espaço para 16 blocos de pelo menos /28."
    }
  }
}
resource "aws_subnet" "private" {
  for_each                = local.subnets
  vpc_id                  = aws_vpc.this.id
  availability_zone       = each.value.az
  cidr_block              = each.value.cidr
  map_public_ip_on_launch = false
  tags                    = merge({ Name = "${var.name}-${each.key}", Tier = each.value.group }, startswith(each.value.group, "web-") ? { "kubernetes.io/role/internal-elb" = "1" } : {})
}
resource "aws_route_table" "private" {
  for_each = toset(local.groups)
  vpc_id   = aws_vpc.this.id
  tags     = { Name = "${var.name}-${each.key}" }
}
resource "aws_route_table_association" "private" {
  for_each       = local.subnets
  subnet_id      = aws_subnet.private[each.key].id
  route_table_id = aws_route_table.private[each.value.group].id
}
resource "aws_subnet" "egress" {
  count             = var.config.nat.enabled ? 1 : 0
  vpc_id            = aws_vpc.this.id
  availability_zone = var.azs[0]
  cidr_block        = cidrsubnet(var.config.vpc.cidr, 4, 12)
  tags              = { Name = "${var.name}-egress" }
}
resource "aws_internet_gateway" "this" {
  count  = var.config.nat.enabled ? 1 : 0
  vpc_id = aws_vpc.this.id
}
resource "aws_route_table" "egress" {
  count  = var.config.nat.enabled ? 1 : 0
  vpc_id = aws_vpc.this.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this[0].id
  }
}
resource "aws_route_table_association" "egress" {
  count          = var.config.nat.enabled ? 1 : 0
  subnet_id      = aws_subnet.egress[0].id
  route_table_id = aws_route_table.egress[0].id
}
resource "aws_eip" "nat" {
  count  = var.config.nat.enabled ? 1 : 0
  domain = "vpc"
}
resource "aws_nat_gateway" "this" {
  count         = var.config.nat.enabled ? 1 : 0
  subnet_id     = aws_subnet.egress[0].id
  allocation_id = aws_eip.nat[0].id
  depends_on    = [aws_internet_gateway.this]
}
resource "aws_route" "web" {
  for_each               = var.config.nat.enabled ? toset(["web-pci", "web-nonpci"]) : toset([])
  route_table_id         = aws_route_table.private[each.key].id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.this[0].id
}
data "aws_region" "current" {}
resource "aws_vpc_endpoint" "s3" {
  count           = var.config.endpoints.s3.enabled ? 1 : 0
  vpc_id          = aws_vpc.this.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [for group in ["web-pci", "web-nonpci"] : aws_route_table.private[group].id]
}
resource "aws_security_group" "endpoints" {
  count       = var.config.endpoints.ecr.enabled ? 1 : 0
  name_prefix = "${var.name}-vpce-"
  vpc_id      = aws_vpc.this.id
  ingress {
    protocol    = "tcp"
    from_port   = 443
    to_port     = 443
    cidr_blocks = [for s in local.subnets : s.cidr if startswith(s.group, "web-")]
  }
}
resource "aws_vpc_endpoint" "ecr" {
  for_each            = var.config.endpoints.ecr.enabled ? toset(["ecr.api", "ecr.dkr"]) : toset([])
  vpc_id              = aws_vpc.this.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.${each.key}"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
  security_group_ids  = [aws_security_group.endpoints[0].id]
  subnet_ids          = [for az in var.azs : aws_subnet.private["web-nonpci-${az}"].id]
}
output "vpc_id" { value = aws_vpc.this.id }
output "subnets" {
  value = { for group in local.groups : group => [for az in var.azs : aws_subnet.private["${group}-${az}"].id] }
}

output "vpc_id" { value = aws_vpc.this.id }
output "subnets" {
  depends_on = [aws_route.web, aws_route_table_association.private]

  value = { for group in local.groups : group => [for az in var.azs : aws_subnet.private["${group}-${az}"].id] }
}

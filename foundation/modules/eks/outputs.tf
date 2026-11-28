output "clusters" {
  value = { for k, cluster in module.cluster : k => { name = cluster.cluster_name, endpoint = cluster.cluster_endpoint } }
}

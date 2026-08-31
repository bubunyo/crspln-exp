output "cluster_name" {
  value = google_container_cluster.primary.name
}

output "cluster_zone" {
  value = google_container_cluster.primary.location
}

output "cluster_endpoint" {
  value     = google_container_cluster.primary.endpoint
  sensitive = true
}

output "get_credentials_command" {
  description = "Run this to configure kubectl for the cluster."
  value       = "gcloud container clusters get-credentials ${google_container_cluster.primary.name} --zone ${google_container_cluster.primary.location} --project ${var.project_id}"
}

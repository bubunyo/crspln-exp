resource "google_container_cluster" "primary" {
  name     = var.cluster_name
  location = var.zone

  remove_default_node_pool = true
  initial_node_count       = 1

  deletion_protection = false

  release_channel {
    channel = "REGULAR"
  }

  ip_allocation_policy {}

  logging_service    = var.enable_logging_monitoring ? "logging.googleapis.com/kubernetes" : "none"
  monitoring_service = var.enable_logging_monitoring ? "monitoring.googleapis.com/kubernetes" : "none"

  addons_config {
    horizontal_pod_autoscaling {
      disabled = true
    }
    http_load_balancing {
      disabled = true
    }
  }
}

resource "google_container_node_pool" "primary_nodes" {
  name     = "${var.cluster_name}-pool"
  cluster  = google_container_cluster.primary.name
  location = var.zone

  node_count = var.node_count

  node_config {
    machine_type    = var.machine_type
    disk_size_gb    = var.disk_size_gb
    disk_type       = var.disk_type
    spot            = var.use_spot
    service_account = google_service_account.gke_node.email

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }
}

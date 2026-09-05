resource "google_artifact_registry_repository" "images" {
  location      = var.region
  repository_id = "images"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository_iam_member" "ci_writer" {
  project    = var.project_id
  location   = google_artifact_registry_repository.images.location
  repository = google_artifact_registry_repository.images.repository_id
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${google_service_account.ci.email}"
}

resource "google_service_account" "gke_node" {
  account_id   = "gke-node"
  display_name = "GKE Node"
}

resource "google_artifact_registry_repository_iam_member" "gke_node_reader" {
  project    = var.project_id
  location   = google_artifact_registry_repository.images.location
  repository = google_artifact_registry_repository.images.repository_id
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.gke_node.email}"
}

output "artifact_registry" {
  value = "${google_artifact_registry_repository.images.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.images.repository_id}"
}

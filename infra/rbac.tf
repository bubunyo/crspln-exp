resource "kubernetes_cluster_role_binding_v1" "ci_admin" {
  metadata {
    name = "terraform-ci-admin"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }

  subject {
    kind      = "User"
    name      = google_service_account.ci.email
    api_group = "rbac.authorization.k8s.io"
  }
}

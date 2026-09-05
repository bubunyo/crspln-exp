resource "kubernetes_namespace_v1" "app" {
  metadata {
    name = "app"
  }
}

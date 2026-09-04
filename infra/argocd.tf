resource "helm_release" "argocd" {
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  version          = "10.4.2"
  namespace        = "argocd"
  create_namespace = true

  values = [file("${path.module}/argocd.yaml")]

  depends_on = [google_container_node_pool.primary_nodes]
  timeout    = 6000
}

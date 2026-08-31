variable "project_id" {
  description = "GCP project ID to deploy into."
  type        = string
  default     = "infra-exp-506709"
}

variable "region" {
  description = "GCP region (used for provider default only; cluster itself is zonal)."
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "GCP zone for the zonal GKE cluster. Zonal clusters are the ones eligible for GCP's free cluster-management-fee waiver (one per billing account)."
  type        = string
  default     = "us-central1-a"
}

variable "cluster_name" {
  description = "Name of the GKE cluster."
  type        = string
  default     = "min-cluster"
}

variable "machine_type" {
  description = "Machine type for the single node pool. e2-small is the smallest type comfortable for a real workload; e2-micro is cheaper but very tight on memory once system pods are scheduled."
  type        = string
  default     = "e2-medium"
}

variable "node_count" {
  description = "Fixed number of nodes (no autoscaling, to keep cost predictable and minimal)."
  type        = number
  default     = 1
}

variable "disk_size_gb" {
  description = "Boot disk size per node, in GB. 30GB is roughly the practical minimum for GKE nodes."
  type        = number
  default     = 30
}

variable "disk_type" {
  description = "Boot disk type. pd-standard is cheaper than pd-balanced/pd-ssd."
  type        = string
  default     = "pd-standard"
}

variable "use_spot" {
  description = "Use Spot VMs for the node pool (large discount, but nodes can be preempted at any time)."
  type        = bool
  default     = true
}

variable "enable_logging_monitoring" {
  description = "Whether to send cluster logs/metrics to Cloud Logging/Monitoring. Disabled by default to avoid any log/metric ingestion cost and to leave more headroom on the single small node."
  type        = bool
  default     = false
}

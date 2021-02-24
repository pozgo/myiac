locals {
  
  firewall_source_ranges  = "81.144.154.0/24"
  network_cidr            = "10.254.0.0/16"

  cluster_name            = "${var.project}-${var.environment}"
  applications_pool_name  = "default-pool"
}

provider "google" {
  project     = var.project
  region      = var.cluster_zone
}

//TODO: Find a way to pass those on run.
terraform {
  backend "gcs" {
    bucket = "ozzy-playground-tfstate-dev"
    prefix = "terraform/state/cluster"
  }
}

data "terraform_remote_state" "state" {
  backend = "gcs"
  config = {
    bucket = var.cluster_state_bucket
    prefix = var.state_bucket_prefix
  }
}

resource "google_container_cluster" "cluster" {
  name     = local.cluster_name
  project  = var.project
  location = var.cluster_zone
  network  = google_compute_network.gke_network.self_link
  subnetwork = google_compute_subnetwork.gke_subnet.self_link
  min_master_version = var.k8s_master_version
  node_version = var.k8s_node_pool_version

  remove_default_node_pool = true
  logging_service          = "logging.googleapis.com/kubernetes"
  monitoring_service       = "monitoring.googleapis.com/kubernetes"
  initial_node_count       = 1

  master_auth {
    username = ""
    password = ""

    client_certificate_config {
      issue_client_certificate = true
    }
  }
}

resource "google_container_node_pool" "applications_node_pool" {
  provider           = google
  project            = var.project
  name               = local.applications_pool_name
  location           = google_container_cluster.cluster.location
  cluster            = google_container_cluster.cluster.name
  initial_node_count = 2

  autoscaling {
    # Minimum number of nodes in the NodePool. Must be >=0 and <= max_node_count.
    min_node_count = 2

    # Maximum number of nodes in the NodePool. Must be >= min_node_count.
    max_node_count = var.applications_max_node_count
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    preemptible  = true
    machine_type = var.applications_machine_type
    disk_size_gb = 10

    oauth_scopes = [
      "https://www.googleapis.com/auth/compute",
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }

  timeouts {
    update = "10m"
  }
}

# Firewall rule for NodePort
resource "google_compute_subnetwork" "gke_subnet" {
  project       = var.project
  name          = "${var.project}-${var.environment}-gke-subnet"
  ip_cidr_range = local.network_cidr
  region        = "europe-west2"
  network       = google_compute_network.gke_network.self_link
}

resource "google_compute_network" "gke_network" {
  project = var.project
  name = "${var.project}-${var.environment}-gke-network"
  auto_create_subnetworks = false
}

resource "google_compute_firewall" "gke_nodeport_service_rule" {
  project = var.project
  name    = "gke-nodeport-firewall-rule"
  network = google_compute_network.gke_network.name

  allow {
    protocol = "tcp"
    ports    = ["30000-32767"]
  }

  source_ranges = [local.firewall_source_ranges]
}

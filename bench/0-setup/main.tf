# terraform {
#   backend "gcs" {
#     bucket = "your-bucket"
#     prefix = "zanzigo-bench"
#   }
# }

locals {
  primary_region = keys(var.regions)[0]
}

provider "google" {
  project = var.project
  region  = local.primary_region
}

provider "google-beta" {
  project = var.project
  region  = local.primary_region
}

resource "google_project_service" "services" {
  for_each = toset([
    "compute.googleapis.com",
    "servicenetworking.googleapis.com",
  ])
  project = var.project
  service = each.value
}

resource "google_compute_network" "network" {
  name                    = "bench"
  auto_create_subnetworks = false

  depends_on = [google_project_service.services]
}

resource "google_compute_firewall" "firewall_iap_ssh" {
  name    = "allow-iap-ssh"
  network = google_compute_network.network.name

  source_ranges = ["35.235.240.0/20"]

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}

resource "google_compute_subnetwork" "subnetworks" {
  for_each = var.regions
  name     = "bench-${each.key}"
  network  = google_compute_network.network.id
  region   = each.key

  private_ip_google_access   = true
  private_ipv6_google_access = "ENABLE_OUTBOUND_VM_ACCESS_TO_GOOGLE"
  ip_cidr_range              = each.value
  purpose                    = "PRIVATE"
}

resource "google_compute_router" "router" {
  for_each = var.regions
  name     = "bench-router-${each.key}"
  network  = google_compute_network.network.id
  region   = each.key
}

resource "google_compute_router_nat" "router_nat" {
  for_each = var.regions
  name     = "bench-nat-${each.key}"
  router   = google_compute_router.router[each.key].name
  region   = each.key

  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

resource "google_compute_global_address" "services_private_ips" {
  name          = "services-private-ips"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.network.id
}

resource "google_service_networking_connection" "services_private" {
  network                 = google_compute_network.network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.services_private_ips.name]
}

resource "google_compute_network_peering_routes_config" "services_private" {
  peering              = google_service_networking_connection.services_private.peering
  network              = google_compute_network.network.name
  import_custom_routes = true
  export_custom_routes = true
}


resource "google_sql_database_instance" "cloudsql_private" {
  name             = "cloudsql-private"
  region           = local.primary_region
  database_version = "POSTGRES_14"

  settings {
    # tier = "db-f1-micro"
    tier = "db-custom-4-8192"
    ip_configuration {
      ipv4_enabled    = "false"
      private_network = google_compute_network.network.id
    }
  }

  deletion_protection = false
}

resource "google_compute_instance" "zanzigo" {
  name         = "zanzigo-${local.primary_region}"
  machine_type = "n2-standard-4"
  zone         = "${local.primary_region}-b" # hardcoded :|

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnetworks[local.primary_region].id
  }

  shielded_instance_config {
    enable_secure_boot = true
  }
}

resource "google_compute_instance" "bench" {
  for_each     = var.regions
  name         = "bench-${each.key}"
  machine_type = "n2-standard-4"
  zone         = "${each.key}-c" # hardcoded :|

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnetworks[each.key].id
  }

  shielded_instance_config {
    enable_secure_boot = true
  }
}

resource "google_compute_firewall" "firewall_allow_zanzigo" {
  name    = "allow-zanzigo"
  network = google_compute_network.network.name

  source_ranges = [for k, v in var.regions : v]

  allow {
    protocol = "tcp"
    ports    = ["4000"]
  }
}



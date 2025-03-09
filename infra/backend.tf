# This backend must be manually bootstrapped in DigitalOcean Spaces.
# Configuration script for this in particular TBD but this will be necessary
# to run once to create the initial state file if starting from scratch.
terraform {
  backend "s3" {
    bucket         = "terraform-state-${var.environment}"
    key            = "${var.environment}/terraform.tfstate"
    region         = "nyc3"
    endpoint       = "https://nyc3.digitaloceanspaces.com"
    access_key     = var.do_spaces_access_key
    secret_key     = var.do_spaces_secret_key
    skip_region_validation      = true
    skip_credentials_validation = true
  }
}

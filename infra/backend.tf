# This backend must be manually bootstrapped in DigitalOcean Spaces.
# Configuration script for this in particular TBD but this will be necessary
# to run once to create the initial remote state file if starting from scratch.

# Most values empty since we will pass these in via CLI (partial config)
# See tf_init.ps1
terraform {
  required_version = ">= 1.11.0"

  backend "s3" {
    endpoints = {
      s3 = "https://nyc3.digitaloceanspaces.com"
    }

    bucket         = ""
    key            = ""
    region         = "nyc3"
    access_key     = ""
    secret_key     = ""

    # These may be AWS specific, so disabling them for DO
    skip_credentials_validation = true
    skip_requesting_account_id  = true
    skip_metadata_api_check     = true
    skip_region_validation      = true
    skip_s3_checksum            = true
  }
}

# Set the variable value in *.tfvars file
# or using -var="do_token=..." CLI option
variable "do_management_token" {
  type        = string
  description = "DigitalOcean API token for infrastructure management"
  sensitive   = true
}

# https://docs.digitalocean.com/platform/regional-availability/
variable "do_region" {
  type        = string
  description = "Region for the resources"
  default     = "nyc1"
}

variable "environment" {
  type        = string
  description = "Environment for the resources (e.g., dev, prod)"
}

variable "github_repo_url" {
  type        = string
  description = "GitHub repository URL for the app"
}

variable "github_repo_branch" {
  type        = string
  description = "Branch to deploy from"
}
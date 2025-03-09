# Set the variable value in *.tfvars file
# or using -var="do_token=..." CLI option
variable "do_token" {
  type        = string
  description = "DigitalOcean API token for infrastructure management"
}

variable "do_region" {
  type        = string
  description = "Region for the resources"
}

variable "environment" {
  type        = string
  description = "Environment for the resources (e.g., dev, prod)"
}

variable "do_spaces_access_key" {
  type        = string
  description = "DigitalOcean Spaces access key"
}

variable "do_spaces_secret_key" {
  type        = string
  description = "DigitalOcean Spaces secret key"
}

variable "github_repo_url" {
  type        = string
  description = "GitHub repository URL for the app"
}

variable "github_repo_branch" {
  type        = string
  description = "Branch to deploy from"
}
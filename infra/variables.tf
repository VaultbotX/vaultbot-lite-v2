# https://cloud.digitalocean.com/account/api/tokens
variable "do_management_token" {
  type        = string
  description = "DigitalOcean API token for infrastructure management"
  sensitive   = true
}

# https://docs.digitalocean.com/platform/regional-availability
variable "do_region" {
  type        = string
  description = "Region for the resources"
  default     = "nyc1"
}

variable "environment" {
  type        = string
  description = "Environment for the resources (e.g., dev, prod)"
}

variable "github_repo" {
  type        = string
  description = "GitHub repository URL for the app"
}

variable "github_repo_branch" {
  type        = string
  description = "Branch to deploy from"
}

variable "discord_token" {
  type        = string
  description = "Discord bot token"
  sensitive   = true
}

variable "discord_administrator_user_id" {
  type        = string
  description = "Discord bot owner user ID"
  sensitive   = true
}

variable "spotify_playlist_id" {
  type        = string
  description = "Spotify playlist ID for the dynamic playlist"
  sensitive   = true
}

variable "genre_spotify_playlist_id" {
  type        = string
  description = "Spotify playlist ID for the genre-based playlist"
  sensitive   = true
}

variable "spotify_client_id" {
  type        = string
  description = "Spotify client ID"
  sensitive   = true
}

variable "spotify_client_secret" {
  type        = string
  description = "Spotify client secret"
  sensitive   = true
}

variable "spotify_token" {
  type        = string
  description = "Spotify oauth token"
  sensitive   = true
}

variable "app_version" {
  type        = string
  description = "Version of the app"
}
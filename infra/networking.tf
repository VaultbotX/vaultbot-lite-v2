resource "digitalocean_vpc" "vaultbot_vpc" {
  name   = "vaultbot-vpc-${var.environment}"
  region = var.do_region
}

# TODO: jump host droplet and firewall to have the ability to securely remote in if necessary
# https://docs.digitalocean.com/products/networking/vpc/how-to/configure-droplet-as-gateway/
# https://docs.digitalocean.com/products/networking/firewalls/getting-started/quickstart/
#!/bin/bash
set -e

echo "=== Catalyst AWS Infrastructure Setup ==="
echo ""

# Check if Terraform is installed
if ! command -v terraform &> /dev/null; then
    echo "Error: Terraform is not installed"
    echo "Install it from: https://www.terraform.io/downloads"
    echo ""
    echo "On macOS: brew install terraform"
    exit 1
fi

# Check if terraform.tfvars exists
if [ ! -f terraform.tfvars ]; then
    echo "Error: terraform.tfvars not found"
    echo ""
    echo "Please create it from the example:"
    echo "  cp terraform.tfvars.example terraform.tfvars"
    echo "  vim terraform.tfvars"
    echo ""
    echo "You need to set:"
    echo "  - ssh_key_name: Your AWS EC2 key pair name"
    echo "  - your_ip: Your IP address (find at https://whatismyip.com)"
    exit 1
fi

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Plan
echo ""
echo "Planning infrastructure..."
terraform plan -out=tfplan

# Confirm
echo ""
echo "This will create:"
echo "  - EC2 instance (t3.medium) in us-east-1"
echo "  - Elastic IP"
echo "  - Security groups"
echo "  - Nginx + systemd configured"
echo ""
read -p "Deploy infrastructure? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Deployment cancelled"
    exit 0
fi

# Apply
echo ""
echo "Deploying infrastructure..."
terraform apply tfplan

# Save outputs
echo ""
echo "Saving connection info..."
terraform output -json > outputs.json

PUBLIC_IP=$(terraform output -raw public_ip)
SSH_KEY=$(grep ssh_key_name terraform.tfvars | cut -d'"' -f2)

echo ""
echo "=== Infrastructure Deployed ==="
echo ""
echo "Public IP: $PUBLIC_IP"
echo "SSH: ssh -i ~/.ssh/${SSH_KEY}.pem ubuntu@${PUBLIC_IP}"
echo ""
echo "Next: Update DNS to point ads.thenexusengine.com to $PUBLIC_IP"
echo ""
echo "Full instructions saved to outputs.json"

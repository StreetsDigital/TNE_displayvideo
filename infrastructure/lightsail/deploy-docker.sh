#!/bin/bash
set -e

if [ -z "$1" ]; then
    if [ -f .lightsail-ip ]; then
        SERVER_IP=$(cat .lightsail-ip)
    else
        echo "Usage: $0 <server-ip>"
        echo "Or run ./deploy.sh first to create instance"
        exit 1
    fi
else
    SERVER_IP=$1
fi

SSH_KEY="~/.ssh/lightsail-us-east-1.pem"
SSH_USER="ec2-user"

echo "=== Deploying Catalyst with Docker to Lightsail ==="
echo ""
echo "Server: $SERVER_IP"
echo "SSH Key: $SSH_KEY"
echo ""

# Test SSH connection
echo "Testing SSH connection..."
ssh -i $SSH_KEY -o StrictHostKeyChecking=no -o ConnectTimeout=10 $SSH_USER@$SERVER_IP "echo '✓ SSH connection successful'" || {
    echo "Error: Cannot connect to server"
    echo "Make sure the instance is running and you have the correct IP"
    exit 1
}

echo ""
echo "Installing Docker on server..."
ssh -i $SSH_KEY $SSH_USER@$SERVER_IP << 'REMOTE_INSTALL'
set -e

# Update system
sudo yum update -y

# Install Docker
sudo yum install -y docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -a -G docker ec2-user

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify
docker --version
docker-compose --version

echo "✓ Docker installed"
REMOTE_INSTALL

echo ""
echo "Creating deployment directory..."
ssh -i $SSH_KEY $SSH_USER@$SERVER_IP "mkdir -p ~/catalyst"

echo ""
echo "Uploading Catalyst files..."
cd ../../  # Back to project root

# Build Docker image locally first (faster than building on Lightsail)
echo "Building Docker image locally..."
docker build -t catalyst-server:latest .

# Save image
echo "Saving Docker image..."
docker save catalyst-server:latest | gzip > /tmp/catalyst-image.tar.gz

# Upload image and files
echo "Uploading to server..."
scp -i $SSH_KEY /tmp/catalyst-image.tar.gz $SSH_USER@$SERVER_IP:~/catalyst/
scp -i $SSH_KEY docker-compose.yml $SSH_USER@$SERVER_IP:~/catalyst/
scp -i $SSH_KEY -r nginx $SSH_USER@$SERVER_IP:~/catalyst/
scp -i $SSH_KEY -r config $SSH_USER@$SERVER_IP:~/catalyst/
scp -i $SSH_KEY -r assets $SSH_USER@$SERVER_IP:~/catalyst/

rm /tmp/catalyst-image.tar.gz

echo ""
echo "Starting Catalyst on server..."
ssh -i $SSH_KEY $SSH_USER@$SERVER_IP << 'REMOTE_START'
set -e
cd ~/catalyst

# Load Docker image
echo "Loading Docker image..."
docker load < catalyst-image.tar.gz
rm catalyst-image.tar.gz

# Start services
echo "Starting Docker containers..."
docker-compose up -d

# Wait for startup
sleep 5

# Check status
echo ""
echo "Container status:"
docker-compose ps

# Test health
echo ""
echo "Testing health endpoint..."
sleep 3
curl -f http://localhost:8000/health && echo "✓ Health check passed" || echo "⚠ Health check failed"

echo ""
echo "Checking logs..."
docker-compose logs --tail=20 catalyst

REMOTE_START

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Services running at:"
echo "  http://$SERVER_IP:8000/health"
echo "  http://$SERVER_IP/health (via Nginx)"
echo ""
echo "Useful commands:"
echo "  ssh -i $SSH_KEY $SSH_USER@$SERVER_IP"
echo "  docker-compose -f ~/catalyst/docker-compose.yml logs -f"
echo "  docker-compose -f ~/catalyst/docker-compose.yml restart"
echo ""
echo "Next: Setup SSL after DNS is pointing to $SERVER_IP"
echo "  ssh -i $SSH_KEY $SSH_USER@$SERVER_IP"
echo "  cd ~/catalyst"
echo "  ./setup-ssl.sh"

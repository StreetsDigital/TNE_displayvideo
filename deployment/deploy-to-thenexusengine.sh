#!/bin/bash
# TNE Catalyst Deployment Script
# Deploy to thenexusengine.com
# Version: 1.0.0
# Date: 2026-02-02

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
DOMAIN_MAIN="ads.thenexusengine.com"
DOMAIN_STAGING="staging.thenexusengine.com"
DOMAIN_GRAFANA="grafana.thenexusengine.com"
DOMAIN_PROMETHEUS="prometheus.thenexusengine.com"
EMAIL="ops@thenexusengine.io"
DEPLOYMENT_DIR="/opt/catalyst"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${MAGENTA}        TNE Catalyst Deployment Script${NC}"
echo -e "${CYAN}        Deploy to thenexusengine.com${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Function to print section header
print_section() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Function to check command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to ask yes/no question
ask_yes_no() {
    local prompt="$1"
    local default="${2:-n}"

    if [ "$default" = "y" ]; then
        prompt="$prompt [Y/n]: "
    else
        prompt="$prompt [y/N]: "
    fi

    while true; do
        read -p "$prompt" response
        response=${response:-$default}
        case "$response" in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}⚠️  This script should be run as root (or with sudo)${NC}"
    if ! ask_yes_no "Continue anyway?"; then
        exit 1
    fi
fi

print_section "Step 1: Prerequisites Check"

echo "Checking required tools..."
MISSING_TOOLS=()

if ! command_exists docker; then
    MISSING_TOOLS+=("docker")
fi

if ! command_exists docker-compose || ! command_exists docker compose; then
    MISSING_TOOLS+=("docker-compose")
fi

if ! command_exists git; then
    MISSING_TOOLS+=("git")
fi

if ! command_exists certbot; then
    MISSING_TOOLS+=("certbot")
fi

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo -e "${RED}❌ Missing required tools: ${MISSING_TOOLS[*]}${NC}"
    echo ""
    echo "Install missing tools:"
    echo "  Ubuntu/Debian:"
    echo "    apt update"
    echo "    apt install -y docker.io docker-compose git certbot python3-certbot-nginx"
    echo ""
    echo "  macOS:"
    echo "    brew install docker docker-compose git certbot"
    echo ""
    exit 1
else
    echo -e "${GREEN}✅ All required tools installed${NC}"
fi

print_section "Step 2: DNS Configuration Check"

echo "Checking DNS records for:"
echo "  - $DOMAIN_MAIN"
echo "  - $DOMAIN_STAGING"
echo "  - $DOMAIN_GRAFANA"
echo "  - $DOMAIN_PROMETHEUS"
echo ""

# Get server IP
SERVER_IP=$(curl -s ifconfig.me || echo "unknown")
echo -e "${CYAN}Server IP: ${SERVER_IP}${NC}"
echo ""

echo "Please ensure the following DNS A records are configured:"
echo ""
echo "  ads.thenexusengine.com         →  ${SERVER_IP}"
echo "  staging.thenexusengine.com     →  ${SERVER_IP}"
echo "  grafana.thenexusengine.com     →  ${SERVER_IP}"
echo "  prometheus.thenexusengine.com  →  ${SERVER_IP}"
echo ""

if ! ask_yes_no "Have you configured DNS records?" "n"; then
    echo -e "${YELLOW}Please configure DNS first, then run this script again.${NC}"
    exit 1
fi

# Verify DNS resolution
echo ""
echo "Verifying DNS resolution..."
for domain in "$DOMAIN_MAIN" "$DOMAIN_STAGING" "$DOMAIN_GRAFANA" "$DOMAIN_PROMETHEUS"; do
    if host "$domain" >/dev/null 2>&1; then
        RESOLVED_IP=$(host "$domain" | grep "has address" | awk '{print $4}' | head -1)
        if [ "$RESOLVED_IP" = "$SERVER_IP" ]; then
            echo -e "${GREEN}✅ $domain → $RESOLVED_IP${NC}"
        else
            echo -e "${YELLOW}⚠️  $domain → $RESOLVED_IP (expected: $SERVER_IP)${NC}"
        fi
    else
        echo -e "${RED}❌ $domain - DNS not resolving${NC}"
    fi
done

echo ""
if ! ask_yes_no "Continue with deployment?" "y"; then
    exit 1
fi

print_section "Step 3: Deployment Directory Setup"

echo "Setting up deployment directory: $DEPLOYMENT_DIR"

if [ -d "$DEPLOYMENT_DIR" ]; then
    echo -e "${YELLOW}⚠️  Directory $DEPLOYMENT_DIR already exists${NC}"
    if ask_yes_no "Backup and recreate?"; then
        BACKUP_DIR="${DEPLOYMENT_DIR}.backup.$(date +%Y%m%d_%H%M%S)"
        mv "$DEPLOYMENT_DIR" "$BACKUP_DIR"
        echo -e "${GREEN}✅ Backed up to $BACKUP_DIR${NC}"
    fi
fi

mkdir -p "$DEPLOYMENT_DIR"
cd "$DEPLOYMENT_DIR"

# Copy deployment files
echo "Copying deployment files..."
cp -r "$SCRIPT_DIR"/* "$DEPLOYMENT_DIR/"
echo -e "${GREEN}✅ Deployment files copied${NC}"

print_section "Step 4: SSL Certificate Setup"

echo "Setting up SSL certificates with Let's Encrypt..."
echo ""
echo -e "${CYAN}Email for SSL notifications: $EMAIL${NC}"
echo ""

# Create SSL directory
mkdir -p ssl

if ask_yes_no "Obtain SSL certificates now?" "y"; then
    echo ""
    echo "Obtaining certificates for all domains..."

    # Stop any running nginx to free port 80
    docker-compose down 2>/dev/null || true

    for domain in "$DOMAIN_MAIN" "$DOMAIN_STAGING" "$DOMAIN_GRAFANA" "$DOMAIN_PROMETHEUS"; do
        echo ""
        echo -e "${CYAN}Obtaining certificate for $domain...${NC}"

        if certbot certonly --standalone \
            --non-interactive \
            --agree-tos \
            --email "$EMAIL" \
            -d "$domain"; then
            echo -e "${GREEN}✅ Certificate obtained for $domain${NC}"

            # Copy certificates to deployment directory
            cp "/etc/letsencrypt/live/$domain/fullchain.pem" "ssl/${domain}.crt"
            cp "/etc/letsencrypt/live/$domain/privkey.pem" "ssl/${domain}.key"
        else
            echo -e "${RED}❌ Failed to obtain certificate for $domain${NC}"
            echo "    Make sure DNS is properly configured and port 80 is accessible"
        fi
    done

    # Create symlinks for default cert
    ln -sf "${DOMAIN_MAIN}.crt" ssl/fullchain.pem
    ln -sf "${DOMAIN_MAIN}.key" ssl/privkey.pem

    echo -e "${GREEN}✅ SSL certificates configured${NC}"
else
    echo -e "${YELLOW}⚠️  Skipping SSL setup. You'll need to configure certificates manually.${NC}"
    echo "    Place certificates in: $DEPLOYMENT_DIR/ssl/"
fi

print_section "Step 5: Environment Configuration"

if [ ! -f ".env" ]; then
    echo "Creating .env from .env.production..."
    cp .env.production .env
    echo -e "${GREEN}✅ Created .env file${NC}"
else
    echo -e "${YELLOW}⚠️  .env file already exists${NC}"
fi

# Verify secrets are set
echo ""
echo "Verifying secrets configuration..."
if grep -q "CHANGE_ME" .env; then
    echo -e "${RED}❌ Found CHANGE_ME placeholder values in .env${NC}"
    echo ""
    if ask_yes_no "Generate new secrets now?" "y"; then
        ./generate-secrets.sh
    else
        echo -e "${YELLOW}⚠️  Please update secrets manually before deployment${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✅ Secrets configured${NC}"
fi

print_section "Step 6: Database Setup"

echo "Setting up database..."

if [ -d "migrations" ] && [ -n "$(ls -A migrations/*.sql 2>/dev/null)" ]; then
    echo -e "${GREEN}✅ Database migrations found: $(ls migrations/*.sql | wc -l) files${NC}"
else
    echo -e "${YELLOW}⚠️  No database migrations found${NC}"
fi

print_section "Step 7: Docker Deployment"

echo "Deploying services with Docker Compose..."
echo ""
echo "Deployment mode:"
echo "  [1] Standard deployment (100% production)"
echo "  [2] Traffic splitting (95% prod, 5% staging)"
echo ""

read -p "Select deployment mode [1]: " DEPLOY_MODE
DEPLOY_MODE=${DEPLOY_MODE:-1}

if [ "$DEPLOY_MODE" = "2" ]; then
    COMPOSE_FILE="docker-compose-split.yml"
    echo -e "${CYAN}Using traffic splitting deployment${NC}"
else
    COMPOSE_FILE="docker-compose.yml"
    echo -e "${CYAN}Using standard deployment${NC}"
fi

# Pull latest images
echo ""
echo "Pulling latest Docker images..."
docker-compose -f "$COMPOSE_FILE" pull

# Start services
echo ""
echo "Starting services..."
docker-compose -f "$COMPOSE_FILE" up -d

echo -e "${GREEN}✅ Services started${NC}"

print_section "Step 8: Health Checks"

echo "Waiting for services to be ready..."
sleep 10

# Check service health
echo ""
echo "Checking service health..."

SERVICES=("catalyst" "postgres" "redis" "nginx")
ALL_HEALTHY=true

for service in "${SERVICES[@]}"; do
    if docker-compose -f "$COMPOSE_FILE" ps | grep -q "$service.*Up"; then
        echo -e "${GREEN}✅ $service is running${NC}"
    else
        echo -e "${RED}❌ $service is not running${NC}"
        ALL_HEALTHY=false
    fi
done

# Test HTTP endpoints
echo ""
echo "Testing HTTP endpoints..."

sleep 5  # Give nginx time to start

if curl -f -k "https://$DOMAIN_MAIN/health" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Health endpoint responding${NC}"
else
    echo -e "${YELLOW}⚠️  Health endpoint not responding (might need DNS propagation)${NC}"
fi

print_section "Step 9: Monitoring Setup"

if ask_yes_no "Deploy monitoring (Grafana/Prometheus)?" "y"; then
    echo ""
    echo "Deploying monitoring stack..."

    cd ../grafana 2>/dev/null || cd grafana 2>/dev/null || echo "Grafana directory not found"

    if [ -f "docker-compose.yml" ]; then
        docker-compose up -d
        echo -e "${GREEN}✅ Monitoring deployed${NC}"
        echo ""
        echo "Access Grafana at: https://$DOMAIN_GRAFANA"
        echo "Default credentials: admin/admin"
    else
        echo -e "${YELLOW}⚠️  Grafana docker-compose.yml not found${NC}"
    fi

    cd "$DEPLOYMENT_DIR"
else
    echo -e "${YELLOW}⚠️  Skipping monitoring setup${NC}"
fi

print_section "Step 10: Firewall Configuration"

echo "Recommended firewall rules:"
echo ""
echo "  sudo ufw allow 22/tcp    # SSH"
echo "  sudo ufw allow 80/tcp    # HTTP"
echo "  sudo ufw allow 443/tcp   # HTTPS"
echo "  sudo ufw enable"
echo ""

if ask_yes_no "Configure firewall now?" "n"; then
    if command_exists ufw; then
        ufw allow 22/tcp
        ufw allow 80/tcp
        ufw allow 443/tcp
        ufw --force enable
        echo -e "${GREEN}✅ Firewall configured${NC}"
    else
        echo -e "${YELLOW}⚠️  ufw not installed${NC}"
    fi
fi

print_section "Deployment Complete! 🚀"

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}        Deployment Successful!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${CYAN}🌐 Service URLs:${NC}"
echo "   Main API:     https://$DOMAIN_MAIN"
echo "   Staging API:  https://$DOMAIN_STAGING"
echo "   Grafana:      https://$DOMAIN_GRAFANA"
echo "   Prometheus:   https://$DOMAIN_PROMETHEUS"
echo ""
echo -e "${CYAN}📊 Quick Health Check:${NC}"
echo "   curl https://$DOMAIN_MAIN/health"
echo ""
echo -e "${CYAN}📋 View Logs:${NC}"
echo "   docker-compose -f $COMPOSE_FILE logs -f"
echo ""
echo -e "${CYAN}🔍 Check Status:${NC}"
echo "   docker-compose -f $COMPOSE_FILE ps"
echo ""
echo -e "${CYAN}📝 Next Steps:${NC}"
echo "   1. Test the API: curl https://$DOMAIN_MAIN/health"
echo "   2. Configure publishers and bidders"
echo "   3. Set up monitoring alerts"
echo "   4. Configure backups (run ./setup-s3-backups.sh)"
echo "   5. Review logs for any errors"
echo ""
echo -e "${CYAN}📖 Documentation:${NC}"
echo "   - Deployment docs: ../docs/deployment/"
echo "   - API Reference: ../docs/api/API-REFERENCE.md"
echo "   - Operations Guide: ../docs/guides/OPERATIONS-GUIDE.md"
echo ""
echo -e "${YELLOW}⚠️  Security Reminders:${NC}"
echo "   - Change Grafana default password (admin/admin)"
echo "   - Store secrets in password manager"
echo "   - Set up SSL auto-renewal: certbot renew --dry-run"
echo "   - Enable backup automation"
echo "   - Review firewall rules"
echo ""
echo -e "${GREEN}Deployment directory: $DEPLOYMENT_DIR${NC}"
echo -e "${GREEN}Configuration file: $DEPLOYMENT_DIR/.env${NC}"
echo -e "${GREEN}SSL certificates: $DEPLOYMENT_DIR/ssl/${NC}"
echo ""

# Save deployment info
cat > DEPLOYMENT_INFO.txt <<EOF
TNE Catalyst Deployment Information
=====================================
Date: $(date)
Server IP: $SERVER_IP
Deployment Directory: $DEPLOYMENT_DIR
Compose File: $COMPOSE_FILE

Domains:
  - Main API: $DOMAIN_MAIN
  - Staging: $DOMAIN_STAGING
  - Grafana: $DOMAIN_GRAFANA
  - Prometheus: $DOMAIN_PROMETHEUS

Contact Email: $EMAIL

Services Deployed:
$(docker-compose -f "$COMPOSE_FILE" ps)

Next SSL Renewal: $(date -d '+90 days' 2>/dev/null || date -v+90d)
EOF

echo -e "${GREEN}✅ Deployment info saved to: DEPLOYMENT_INFO.txt${NC}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

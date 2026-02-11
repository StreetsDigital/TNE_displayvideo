# AWS Infrastructure for Catalyst

This Terraform configuration provisions everything needed to run Catalyst in AWS us-east-1 (closest to major SSPs).

## What Gets Created

- **EC2 Instance:** t3.medium (2 vCPU, 4GB RAM, 30GB disk)
- **Elastic IP:** Static public IP address
- **Security Groups:** Ports 22 (SSH), 80 (HTTP), 443 (HTTPS), 8000 (Catalyst)
- **Ubuntu 22.04:** Latest LTS with Go, Nginx, and Catalyst configured
- **Systemd Service:** Auto-starts Catalyst on boot
- **Nginx:** Reverse proxy with SSL support

**Location:** us-east-1 (N. Virginia) - closest to Excelsior 5 and major SSPs (Magnite, PubMatic, etc.)

## Prerequisites

1. **AWS Account** with EC2 permissions
2. **Terraform** installed (`brew install terraform`)
3. **AWS CLI** configured with credentials
4. **EC2 Key Pair** created in us-east-1 region

### Create EC2 Key Pair (if needed)

```bash
# Via AWS CLI
aws ec2 create-key-pair --region us-east-1 --key-name catalyst-key --query 'KeyMaterial' --output text > ~/.ssh/catalyst-key.pem
chmod 400 ~/.ssh/catalyst-key.pem

# Or via AWS Console:
# EC2 → Key Pairs → Create Key Pair → catalyst-key → Download .pem file
```

## Quick Start

### 1. Configure Variables

```bash
cd infrastructure/aws
cp terraform.tfvars.example terraform.tfvars
vim terraform.tfvars
```

Set these values:
```hcl
ssh_key_name = "catalyst-key"  # Your key pair name
your_ip      = "1.2.3.4/32"    # Your IP (find at https://whatismyip.com)
```

### 2. Deploy Infrastructure

```bash
./deploy.sh
```

This will:
- Initialize Terraform
- Show you what will be created
- Ask for confirmation
- Deploy everything to AWS
- Output the public IP and next steps

### 3. Update DNS

Point `ads.thenexusengine.com` to the Elastic IP:

```
A record: ads.thenexusengine.com → <ELASTIC_IP>
```

Wait 5-10 minutes for DNS propagation.

### 4. Deploy Catalyst

```bash
# From your local machine
cd /Users/andrewstreets/tnevideo

# Upload deployment package
scp -i ~/.ssh/catalyst-key.pem build/catalyst-deployment.tar.gz ubuntu@<ELASTIC_IP>:/tmp/

# SSH to server
ssh -i ~/.ssh/catalyst-key.pem ubuntu@<ELASTIC_IP>

# On the server - extract and deploy
sudo tar xzf /tmp/catalyst-deployment.tar.gz -C /opt/catalyst --strip-components=1
sudo chown -R catalyst:catalyst /opt/catalyst
sudo chmod +x /opt/catalyst/catalyst-server
sudo systemctl start catalyst

# Check status
sudo systemctl status catalyst
curl http://localhost:8000/health
```

### 5. Setup SSL (after DNS propagation)

```bash
# On the server
sudo certbot --nginx -d ads.thenexusengine.com --non-interactive --agree-tos --email your@email.com

# Test HTTPS
curl https://ads.thenexusengine.com/health
```

## Manual Deployment (Alternative)

If you prefer to use Terraform directly:

```bash
cd infrastructure/aws

# Initialize
terraform init

# Plan
terraform plan

# Apply
terraform apply

# Get outputs
terraform output
```

## Infrastructure Details

### Instance Specifications
- **Type:** t3.medium
- **vCPUs:** 2
- **RAM:** 4GB
- **Disk:** 30GB SSD (gp3)
- **Network:** Enhanced networking enabled
- **Region:** us-east-1
- **Estimated Cost:** ~$30/month

### Ports
- 22: SSH (restricted to your IP)
- 80: HTTP (open to world)
- 443: HTTPS (open to world)
- 8000: Catalyst application (open to world)

### Software Pre-installed
- Ubuntu 22.04 LTS
- Go 1.21
- Nginx
- Certbot (for SSL)
- systemd service for Catalyst

## Verification

After deployment, verify everything works:

```bash
# SSH access
ssh -i ~/.ssh/catalyst-key.pem ubuntu@<ELASTIC_IP>

# Server setup complete
cat /opt/catalyst/setup-complete.txt

# Nginx running
sudo systemctl status nginx

# Catalyst service configured
sudo systemctl status catalyst
```

## Updating Infrastructure

To change instance type or other settings:

```bash
# Edit terraform.tfvars
vim terraform.tfvars

# Apply changes
terraform plan
terraform apply
```

## Destroying Infrastructure

To tear down everything:

```bash
terraform destroy
```

**Warning:** This will delete the EC2 instance and Elastic IP. Make sure you have backups!

## Costs

Estimated monthly costs:
- EC2 t3.medium: ~$30
- EBS 30GB gp3: ~$3
- Elastic IP (while attached): $0
- Data transfer: ~$9/TB

**Total:** ~$33-40/month

## Troubleshooting

### Can't connect via SSH
- Check your IP hasn't changed: `curl https://whatismyip.com`
- Update security group: Edit `your_ip` in terraform.tfvars and run `terraform apply`

### Catalyst won't start
```bash
sudo journalctl -u catalyst -n 100
```

### Nginx errors
```bash
sudo nginx -t
sudo systemctl status nginx
sudo journalctl -u nginx
```

### DNS not resolving
- Check DNS propagation: `dig ads.thenexusengine.com`
- Wait 5-10 minutes for DNS to propagate globally

## Security Notes

- SSH is restricted to your IP only
- All services run as non-root user (catalyst)
- Disk encryption enabled
- Automatic security updates enabled
- HTTPS required for production traffic

## Support

For issues:
1. Check AWS Console → EC2 → Instances
2. Review CloudWatch logs
3. SSH to server and check: `sudo journalctl -u catalyst -f`

## Next Steps After Infrastructure is Live

1. Deploy Catalyst binary
2. Setup SSL with Certbot
3. Test health endpoint
4. Run load tests
5. Configure monitoring
6. Update BizBudding with production endpoint

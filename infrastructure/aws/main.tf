terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# Variables
variable "project_name" {
  default = "catalyst-server"
}

variable "instance_type" {
  default = "t3.medium" # 2 vCPU, 4GB RAM
}

variable "ssh_key_name" {
  description = "Name of your AWS EC2 key pair"
  type        = string
}

variable "your_ip" {
  description = "Your IP address for SSH access (CIDR format, e.g., 1.2.3.4/32)"
  type        = string
}

# Data sources
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

# VPC and Networking
resource "aws_default_vpc" "default" {
  tags = {
    Name = "${var.project_name}-vpc"
  }
}

resource "aws_security_group" "catalyst" {
  name        = "${var.project_name}-sg"
  description = "Security group for Catalyst ad server"
  vpc_id      = aws_default_vpc.default.id

  # SSH access
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.your_ip]
    description = "SSH access"
  }

  # HTTP (for health checks, testing)
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP"
  }

  # HTTPS
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS"
  }

  # Catalyst application port
  ingress {
    from_port   = 8000
    to_port     = 8000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Catalyst application"
  }

  # Outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound traffic"
  }

  tags = {
    Name = "${var.project_name}-sg"
  }
}

# Elastic IP
resource "aws_eip" "catalyst" {
  domain = "vpc"
  tags = {
    Name = "${var.project_name}-eip"
  }
}

resource "aws_eip_association" "catalyst" {
  instance_id   = aws_instance.catalyst.id
  allocation_id = aws_eip.catalyst.id
}

# EC2 Instance
resource "aws_instance" "catalyst" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type
  key_name      = var.ssh_key_name

  vpc_security_group_ids = [aws_security_group.catalyst.id]

  root_block_device {
    volume_size = 30 # GB
    volume_type = "gp3"
    encrypted   = true
  }

  user_data = <<-EOF
              #!/bin/bash
              set -e

              # Update system
              apt-get update
              apt-get upgrade -y

              # Install dependencies
              apt-get install -y \
                curl \
                wget \
                git \
                nginx \
                certbot \
                python3-certbot-nginx \
                jq \
                htop

              # Install Go 1.21
              wget -q https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
              rm -rf /usr/local/go
              tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
              echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
              rm go1.21.6.linux-amd64.tar.gz

              # Create catalyst user and directories
              useradd -r -s /bin/bash -d /opt/catalyst catalyst || true
              mkdir -p /opt/catalyst/{config,assets,logs}
              chown -R catalyst:catalyst /opt/catalyst

              # Configure Nginx
              cat > /etc/nginx/sites-available/catalyst << 'NGINX'
              server {
                  listen 80;
                  server_name ads.thenexusengine.com;

                  # Static assets
                  location /assets/ {
                      alias /opt/catalyst/assets/;
                      expires 1h;
                      add_header Cache-Control "public, immutable";
                  }

                  # Test pages
                  location ~ \.(html)$ {
                      root /opt/catalyst/assets;
                  }

                  # Proxy to Catalyst
                  location / {
                      proxy_pass http://localhost:8000;
                      proxy_set_header Host $host;
                      proxy_set_header X-Real-IP $remote_addr;
                      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                      proxy_set_header X-Forwarded-Proto $scheme;
                      proxy_connect_timeout 3s;
                      proxy_send_timeout 3s;
                      proxy_read_timeout 3s;
                  }
              }
              NGINX

              ln -sf /etc/nginx/sites-available/catalyst /etc/nginx/sites-enabled/
              rm -f /etc/nginx/sites-enabled/default
              nginx -t && systemctl restart nginx

              # Create systemd service
              cat > /etc/systemd/system/catalyst.service << 'SERVICE'
              [Unit]
              Description=Catalyst Ad Server
              After=network.target

              [Service]
              Type=simple
              User=catalyst
              Group=catalyst
              WorkingDirectory=/opt/catalyst
              ExecStart=/opt/catalyst/catalyst-server
              Restart=always
              RestartSec=5
              StandardOutput=journal
              StandardError=journal
              SyslogIdentifier=catalyst

              # Environment
              Environment="PORT=8000"
              Environment="HOST_URL=https://ads.thenexusengine.com"
              Environment="LOG_LEVEL=info"
              Environment="PBS_TIMEOUT=2500ms"

              # Limits
              LimitNOFILE=65536
              LimitNPROC=4096

              [Install]
              WantedBy=multi-user.target
              SERVICE

              systemctl daemon-reload
              systemctl enable catalyst

              # Setup complete
              echo "Server setup complete" > /opt/catalyst/setup-complete.txt
              EOF

  tags = {
    Name        = var.project_name
    Environment = "production"
    ManagedBy   = "terraform"
  }
}

# Outputs
output "instance_id" {
  value       = aws_instance.catalyst.id
  description = "EC2 instance ID"
}

output "public_ip" {
  value       = aws_eip.catalyst.public_ip
  description = "Elastic IP address"
}

output "ssh_command" {
  value       = "ssh -i ~/.ssh/${var.ssh_key_name}.pem ubuntu@${aws_eip.catalyst.public_ip}"
  description = "SSH command to connect"
}

output "dns_record" {
  value       = "Create A record: ads.thenexusengine.com -> ${aws_eip.catalyst.public_ip}"
  description = "DNS configuration needed"
}

output "next_steps" {
  value = <<-EOT

    Next steps:
    1. Update DNS: Point ads.thenexusengine.com to ${aws_eip.catalyst.public_ip}
    2. Wait for DNS propagation (5-10 minutes)
    3. Deploy Catalyst:
       cd /Users/andrewstreets/tnevideo
       scp -i ~/.ssh/${var.ssh_key_name}.pem build/catalyst-deployment.tar.gz ubuntu@${aws_eip.catalyst.public_ip}:/tmp/
       ssh -i ~/.ssh/${var.ssh_key_name}.pem ubuntu@${aws_eip.catalyst.public_ip}

    4. On the server:
       sudo tar xzf /tmp/catalyst-deployment.tar.gz -C /opt/catalyst --strip-components=1
       sudo chown -R catalyst:catalyst /opt/catalyst
       sudo chmod +x /opt/catalyst/catalyst-server
       sudo systemctl start catalyst

    5. Setup SSL:
       sudo certbot --nginx -d ads.thenexusengine.com

    6. Test:
       curl https://ads.thenexusengine.com/health
  EOT
}

#!/bin/bash
set -e

echo "=== Catalyst Lightsail Deployment ==="
echo ""

# Configuration
INSTANCE_NAME="catalyst-server"
REGION="us-east-1"
BLUEPRINT="amazon_linux_2"
BUNDLE="medium_2_0"  # 2GB RAM, 2 vCPU, 60GB SSD - $20/month

echo "Configuration:"
echo "  Instance: $INSTANCE_NAME"
echo "  Region: $REGION"
echo "  Plan: $BUNDLE (2GB RAM, $20/month)"
echo ""

# Check if instance already exists
if aws lightsail get-instance --instance-name $INSTANCE_NAME --region $REGION &>/dev/null; then
    echo "Instance '$INSTANCE_NAME' already exists!"
    PUBLIC_IP=$(aws lightsail get-instance --instance-name $INSTANCE_NAME --region $REGION --query 'instance.publicIpAddress' --output text)
    echo "Public IP: $PUBLIC_IP"
    echo ""
    read -p "Use existing instance? (yes/no): " use_existing
    if [ "$use_existing" != "yes" ]; then
        echo "Exiting. Delete the instance first if you want to recreate it:"
        echo "  aws lightsail delete-instance --instance-name $INSTANCE_NAME --region $REGION"
        exit 1
    fi
else
    echo "Creating Lightsail instance..."

    # Create instance
    aws lightsail create-instances \
        --instance-names $INSTANCE_NAME \
        --availability-zone ${REGION}a \
        --blueprint-id $BLUEPRINT \
        --bundle-id $BUNDLE \
        --region $REGION \
        --tags key=Project,value=Catalyst key=Environment,value=Production

    echo "✓ Instance created"
    echo "Waiting for instance to start (this takes ~2 minutes)..."

    # Wait for instance to be running
    while true; do
        STATE=$(aws lightsail get-instance --instance-name $INSTANCE_NAME --region $REGION --query 'instance.state.name' --output text)
        if [ "$STATE" == "running" ]; then
            break
        fi
        echo "  Status: $STATE... waiting"
        sleep 10
    done

    echo "✓ Instance is running"

    # Get public IP
    PUBLIC_IP=$(aws lightsail get-instance --instance-name $INSTANCE_NAME --region $REGION --query 'instance.publicIpAddress' --output text)
    echo "✓ Public IP: $PUBLIC_IP"

    # Open ports
    echo "Opening firewall ports..."
    aws lightsail put-instance-public-ports \
        --instance-name $INSTANCE_NAME \
        --port-infos \
            fromPort=22,toPort=22,protocol=TCP \
            fromPort=80,toPort=80,protocol=TCP \
            fromPort=443,toPort=443,protocol=TCP \
            fromPort=8000,toPort=8000,protocol=TCP \
        --region $REGION

    echo "✓ Firewall configured"

    # Allocate static IP
    echo "Allocating static IP..."
    aws lightsail allocate-static-ip \
        --static-ip-name ${INSTANCE_NAME}-ip \
        --region $REGION 2>/dev/null || echo "Static IP may already exist"

    aws lightsail attach-static-ip \
        --static-ip-name ${INSTANCE_NAME}-ip \
        --instance-name $INSTANCE_NAME \
        --region $REGION

    STATIC_IP=$(aws lightsail get-static-ip --static-ip-name ${INSTANCE_NAME}-ip --region $REGION --query 'staticIp.ipAddress' --output text)
    echo "✓ Static IP: $STATIC_IP"
    PUBLIC_IP=$STATIC_IP

    echo ""
    echo "Waiting 30 seconds for instance to fully initialize..."
    sleep 30
fi

# Get SSH key
echo ""
echo "Downloading SSH key..."
aws lightsail download-default-key-pair --region $REGION --query 'privateKeyBase64' --output text | base64 -d > ~/.ssh/lightsail-$REGION.pem
chmod 400 ~/.ssh/lightsail-$REGION.pem
echo "✓ SSH key saved to ~/.ssh/lightsail-$REGION.pem"

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Instance Details:"
echo "  Name: $INSTANCE_NAME"
echo "  IP: $PUBLIC_IP"
echo "  SSH: ssh -i ~/.ssh/lightsail-$REGION.pem ec2-user@$PUBLIC_IP"
echo ""
echo "Next steps:"
echo "1. Update DNS: ads.thenexusengine.com → $PUBLIC_IP"
echo "2. Deploy Docker: ./deploy-docker.sh $PUBLIC_IP"
echo ""

# Save IP to file
echo $PUBLIC_IP > .lightsail-ip
echo "✓ IP saved to .lightsail-ip"

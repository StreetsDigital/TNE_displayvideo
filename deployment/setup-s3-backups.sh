#!/bin/bash
# AWS S3 Bucket Setup for Production Backups
# Creates S3 bucket with encryption, versioning, and lifecycle policies

set -euo pipefail

# Configuration
BUCKET_NAME="${BACKUP_S3_BUCKET:-catalyst-prod-backups}"
AWS_REGION="${AWS_DEFAULT_REGION:-us-east-1}"
LIFECYCLE_DAYS_TO_GLACIER="${LIFECYCLE_DAYS_TO_GLACIER:-30}"
LIFECYCLE_DAYS_TO_EXPIRE="${LIFECYCLE_DAYS_TO_EXPIRE:-90}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================="
echo "AWS S3 Backup Bucket Setup"
echo "========================================="
echo "Bucket name: ${BUCKET_NAME}"
echo "Region: ${AWS_REGION}"
echo ""

# Check AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo -e "${RED}❌ AWS CLI not installed${NC}"
    echo "Install with: pip install awscli"
    exit 1
fi

# Check AWS credentials configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}❌ AWS credentials not configured${NC}"
    echo "Configure with: aws configure"
    exit 1
fi

echo -e "${GREEN}✅ AWS CLI configured${NC}"
echo "Account: $(aws sts get-caller-identity --query Account --output text)"
echo ""

# Create S3 bucket
echo "Step 1: Creating S3 bucket..."
if aws s3 ls "s3://${BUCKET_NAME}" 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Bucket already exists: ${BUCKET_NAME}${NC}"
else
    if [ "${AWS_REGION}" = "us-east-1" ]; then
        # us-east-1 doesn't need LocationConstraint
        aws s3 mb "s3://${BUCKET_NAME}"
    else
        aws s3 mb "s3://${BUCKET_NAME}" --region "${AWS_REGION}"
    fi
    echo -e "${GREEN}✅ Bucket created: ${BUCKET_NAME}${NC}"
fi

# Enable versioning
echo ""
echo "Step 2: Enabling versioning..."
aws s3api put-bucket-versioning \
    --bucket "${BUCKET_NAME}" \
    --versioning-configuration Status=Enabled
echo -e "${GREEN}✅ Versioning enabled${NC}"

# Enable server-side encryption
echo ""
echo "Step 3: Enabling server-side encryption (AES256)..."
aws s3api put-bucket-encryption \
    --bucket "${BUCKET_NAME}" \
    --server-side-encryption-configuration '{
      "Rules": [
        {
          "ApplyServerSideEncryptionByDefault": {
            "SSEAlgorithm": "AES256"
          },
          "BucketKeyEnabled": true
        }
      ]
    }'
echo -e "${GREEN}✅ Encryption enabled${NC}"

# Block public access
echo ""
echo "Step 4: Blocking public access..."
aws s3api put-public-access-block \
    --bucket "${BUCKET_NAME}" \
    --public-access-block-configuration \
        "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"
echo -e "${GREEN}✅ Public access blocked${NC}"

# Set lifecycle policy
echo ""
echo "Step 5: Configuring lifecycle policy..."
cat > /tmp/lifecycle-policy.json <<EOF
{
  "Rules": [
    {
      "Id": "BackupRetentionPolicy",
      "Status": "Enabled",
      "Filter": {
        "Prefix": "catalyst-backups/"
      },
      "Transitions": [
        {
          "Days": ${LIFECYCLE_DAYS_TO_GLACIER},
          "StorageClass": "GLACIER"
        }
      ],
      "Expiration": {
        "Days": ${LIFECYCLE_DAYS_TO_EXPIRE}
      },
      "NoncurrentVersionExpiration": {
        "NoncurrentDays": 7
      }
    }
  ]
}
EOF

aws s3api put-bucket-lifecycle-configuration \
    --bucket "${BUCKET_NAME}" \
    --lifecycle-configuration file:///tmp/lifecycle-policy.json

rm /tmp/lifecycle-policy.json
echo -e "${GREEN}✅ Lifecycle policy configured${NC}"
echo "   - Transition to Glacier after ${LIFECYCLE_DAYS_TO_GLACIER} days"
echo "   - Delete after ${LIFECYCLE_DAYS_TO_EXPIRE} days"

# Create IAM user for backups
echo ""
echo "Step 6: Creating IAM user for backups..."
IAM_USER="catalyst-backup"

if aws iam get-user --user-name "${IAM_USER}" &> /dev/null; then
    echo -e "${YELLOW}⚠️  IAM user already exists: ${IAM_USER}${NC}"
else
    aws iam create-user --user-name "${IAM_USER}" > /dev/null
    echo -e "${GREEN}✅ IAM user created: ${IAM_USER}${NC}"
fi

# Create IAM policy for bucket access
echo ""
echo "Step 7: Creating IAM policy..."
POLICY_NAME="CatalystBackupPolicy"
POLICY_ARN="arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):policy/${POLICY_NAME}"

cat > /tmp/backup-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::${BUCKET_NAME}",
        "arn:aws:s3:::${BUCKET_NAME}/*"
      ]
    }
  ]
}
EOF

if aws iam get-policy --policy-arn "${POLICY_ARN}" &> /dev/null; then
    echo -e "${YELLOW}⚠️  IAM policy already exists: ${POLICY_NAME}${NC}"
else
    aws iam create-policy \
        --policy-name "${POLICY_NAME}" \
        --policy-document file:///tmp/backup-policy.json > /dev/null
    echo -e "${GREEN}✅ IAM policy created: ${POLICY_NAME}${NC}"
fi

rm /tmp/backup-policy.json

# Attach policy to user
echo ""
echo "Step 8: Attaching policy to user..."
if aws iam list-attached-user-policies --user-name "${IAM_USER}" | grep -q "${POLICY_NAME}"; then
    echo -e "${YELLOW}⚠️  Policy already attached to user${NC}"
else
    aws iam attach-user-policy \
        --user-name "${IAM_USER}" \
        --policy-arn "${POLICY_ARN}"
    echo -e "${GREEN}✅ Policy attached to user${NC}"
fi

# Create access keys
echo ""
echo "Step 9: Creating access keys..."
echo -e "${YELLOW}⚠️  Checking for existing access keys...${NC}"

EXISTING_KEYS=$(aws iam list-access-keys --user-name "${IAM_USER}" --query 'AccessKeyMetadata[*].AccessKeyId' --output text)
if [ -n "${EXISTING_KEYS}" ]; then
    echo -e "${YELLOW}⚠️  User already has access keys:${NC}"
    echo "${EXISTING_KEYS}"
    echo ""
    echo "Delete existing keys? (y/N)"
    read -r response
    if [ "${response}" = "y" ]; then
        for key in ${EXISTING_KEYS}; do
            aws iam delete-access-key --user-name "${IAM_USER}" --access-key-id "${key}"
            echo "   Deleted: ${key}"
        done
    else
        echo "Skipping access key creation."
        exit 0
    fi
fi

# Create new access keys
CREDENTIALS=$(aws iam create-access-key --user-name "${IAM_USER}" --output json)
ACCESS_KEY_ID=$(echo "${CREDENTIALS}" | grep -o '"AccessKeyId": "[^"]*' | cut -d'"' -f4)
SECRET_ACCESS_KEY=$(echo "${CREDENTIALS}" | grep -o '"SecretAccessKey": "[^"]*' | cut -d'"' -f4)

echo -e "${GREEN}✅ Access keys created${NC}"
echo ""
echo "========================================="
echo "IMPORTANT: Save these credentials!"
echo "========================================="
echo ""
echo "Add to your .env.production file:"
echo ""
echo "BACKUP_S3_BUCKET=${BUCKET_NAME}"
echo "AWS_ACCESS_KEY_ID=${ACCESS_KEY_ID}"
echo "AWS_SECRET_ACCESS_KEY=${SECRET_ACCESS_KEY}"
echo "AWS_DEFAULT_REGION=${AWS_REGION}"
echo ""
echo "========================================="
echo ""
echo -e "${RED}⚠️  These credentials will NOT be shown again!${NC}"
echo -e "${RED}Save them now to a secure location.${NC}"
echo ""

# Test bucket access
echo "Step 10: Testing bucket access..."
echo "test" | aws s3 cp - "s3://${BUCKET_NAME}/test.txt"
aws s3 rm "s3://${BUCKET_NAME}/test.txt"
echo -e "${GREEN}✅ Bucket access test successful${NC}"

echo ""
echo "========================================="
echo "S3 Backup Bucket Setup Complete!"
echo "========================================="
echo ""
echo "Summary:"
echo "  Bucket: ${BUCKET_NAME}"
echo "  Region: ${AWS_REGION}"
echo "  Encryption: AES256"
echo "  Versioning: Enabled"
echo "  Lifecycle: ${LIFECYCLE_DAYS_TO_GLACIER}d → Glacier, ${LIFECYCLE_DAYS_TO_EXPIRE}d → Delete"
echo "  IAM User: ${IAM_USER}"
echo ""
echo "Next steps:"
echo "  1. Add credentials to .env.production"
echo "  2. Deploy backup service: docker-compose up -d backup"
echo "  3. Test backup: docker exec catalyst-backup /usr/local/bin/backup-postgres.sh"
echo "  4. Verify S3 upload: aws s3 ls s3://${BUCKET_NAME}/catalyst-backups/"
echo ""

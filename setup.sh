#!/bin/bash

# Quick setup script for AWS Lambda Hello World project
# This script helps you get started with the project setup

set -e

echo "🚀 AWS Lambda Hello World Setup Script"
echo "======================================"

# Check if required tools are installed
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ $1 is not installed. Please install it first."
        exit 1
    else
        echo "✅ $1 is installed"
    fi
}

echo "📋 Checking required tools..."
check_tool "aws"
check_tool "go"
check_tool "terraform"
check_tool "terragrunt"

echo ""
echo "🔧 Project Setup Checklist:"
echo "1. Create an S3 bucket for deployments:"
echo "   aws s3 mb s3://your-deployment-bucket-name"
echo ""
echo "2. Create a DynamoDB table for Terraform state locking:"
echo "   aws dynamodb create-table \\"
echo "     --table-name terraform-locks \\"
echo "     --attribute-definitions AttributeName=LockID,AttributeType=S \\"
echo "     --key-schema AttributeName=LockID,KeyType=HASH \\"
echo "     --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5"
echo ""
echo "3. Update terragrunt/terragrunt.hcl with your S3 bucket names"
echo ""
echo "4. Set up GitHub repository secrets:"
echo "   - AWS_ACCESS_KEY_ID"
echo "   - AWS_SECRET_ACCESS_KEY"
echo ""
echo "5. Set up GitHub repository variables:"
echo "   - S3_DEPLOYMENT_BUCKET"
echo "   - AWS_REGION (optional, defaults to us-east-1)"
echo ""

# Get user input for bucket name
read -p "Enter your S3 deployment bucket name (or press Enter to skip): " bucket_name

if [ ! -z "$bucket_name" ]; then
    echo ""
    echo "🏗️  Building and testing the Lambda function..."
    
    # Build the function
    make build
    
    echo "✅ Lambda function built successfully!"
    echo ""
    echo "🚀 To deploy to dev environment, run:"
    echo "   make upload deploy-dev S3_BUCKET=$bucket_name"
    echo ""
    echo "🧪 To test the deployed function, run:"
    echo "   make test-lambda ENV=dev"
    echo ""
else
    echo ""
    echo "⏭️  Skipping build. You can build manually with:"
    echo "   make build"
fi

echo ""
echo "📖 For more information, see the README.md file"
echo "✨ Happy coding!"
# AWS Lambda Hello World with Go and AWS SDK v2

This project demonstrates a complete AWS Lambda deployment pipeline using:
- **Go** with AWS SDK v2
- **GitHub Actions** for CI/CD
- **Terraform** for infrastructure as code
- **Terragrunt** for environment management

## 🏗️ Project Structure

```
├── src/lambda/              # Go Lambda function source code
├── infrastructure/terraform/lambda/  # Terraform modules
├── terragrunt/             # Environment-specific configurations
│   ├── dev/
│   ├── staging/
│   └── prod/
├── .github/workflows/      # CI/CD pipelines
├── go.mod                  # Go module definition
└── README.md
```

## 🚀 Features

- **Hello World Lambda**: Simple Go function using AWS SDK v2
- **S3 Integration**: Demonstrates SDK usage by listing S3 buckets
- **Automated Build**: GitHub Actions workflow for building and uploading to S3
- **Automated Deploy**: GitHub Actions workflow for Terraform/Terragrunt deployment
- **Multi-Environment**: Support for dev, staging, and prod environments
- **Function URL**: Direct HTTP access to Lambda function

## 📋 Prerequisites

1. **AWS Account** with appropriate permissions
2. **S3 Bucket** for storing deployment packages and Terraform state
3. **GitHub Repository Secrets**:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
4. **GitHub Repository Variables**:
   - `S3_DEPLOYMENT_BUCKET` - S3 bucket for deployment packages
   - `AWS_REGION` - AWS region (defaults to us-east-1)

## 🛠️ Setup Instructions

### 1. AWS Setup

Create an S3 bucket for deployments:
```bash
aws s3 mb s3://your-deployment-bucket
```

Create a DynamoDB table for Terraform state locking:
```bash
aws dynamodb create-table \
  --table-name terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 2. Update Terragrunt Configuration

Edit `terragrunt/terragrunt.hcl` and update:
- `bucket = "your-terraform-state-bucket-${local.environment}"` with your actual bucket name

### 3. GitHub Repository Setup

Configure the following secrets in your GitHub repository:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

Configure the following variables:
- `S3_DEPLOYMENT_BUCKET` - Your S3 deployment bucket name
- `AWS_REGION` - Your preferred AWS region

## 🔄 CI/CD Workflows

### Build and Upload Workflow
- **Trigger**: Push to main/develop or PR to main
- **Actions**:
  - Build Go Lambda function
  - Create deployment package
  - Upload to S3
  - Create deployment manifest

### Deploy Workflow
- **Trigger**: Push to main or manual dispatch
- **Actions**:
  - Deploy infrastructure with Terragrunt
  - Test Lambda function
  - Output function URL and ARN

## 🧪 Local Development

### Build the Lambda function:
```bash
cd src/lambda
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
```

### Test locally (requires AWS credentials):
```bash
go run src/lambda/main.go
```

### Deploy manually with Terragrunt:
```bash
cd terragrunt/dev/lambda
export TF_VAR_lambda_s3_bucket="your-bucket-name"
export TF_VAR_lambda_s3_key="path/to/your/deployment.zip"
terragrunt plan
terragrunt apply
```

## 📝 Lambda Function Details

The Lambda function (`src/lambda/main.go`):
- Uses AWS SDK v2 for AWS service integration
- Handles API Gateway proxy events
- Returns JSON responses
- Demonstrates S3 bucket listing as an SDK example
- Includes proper error handling and logging

## 🌍 Environment Management

The project supports three environments:
- **dev**: Development environment
- **staging**: Staging environment  
- **prod**: Production environment

Each environment has its own:
- Terragrunt configuration
- Resource naming
- Terraform state

## 🔐 Security Features

- **IAM Role**: Least privilege permissions for Lambda
- **Encryption**: Terraform state encryption
- **State Locking**: DynamoDB-based locking
- **Environment Isolation**: Separate resources per environment

## 📊 Monitoring

- **CloudWatch Logs**: Automatic log group creation
- **Function URL**: Easy HTTP testing
- **Deployment Manifests**: Track deployment metadata

## 🚦 Usage

Once deployed, you can test the Lambda function:

1. **Via Function URL**: Visit the URL output from deployment
2. **Via AWS Console**: Test in the Lambda console
3. **Via CLI**:
   ```bash
   aws lambda invoke \
     --function-name hello-world-lambda-dev \
     --payload '{"httpMethod": "GET", "path": "/"}' \
     response.json && cat response.json
   ```

## 🤝 Contributing

1. Create a feature branch
2. Make your changes
3. Push to trigger the build workflow
4. Create a PR to trigger deployment to staging
5. Merge to main for production deployment
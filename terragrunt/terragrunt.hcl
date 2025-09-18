# Terragrunt con# Configure remote state
remote_state {
  backend = "s3"
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
  config = {
    bucket = "integracionbucket"
    key    = "terraform-state/${local.environment}/${local.service}/terraform.tfstate"
    region = "us-east-1"
    encrypt = true
  }
}common settings
locals {
  # Parse the file path to get environment info
  path_parts   = split("/", path_relative_to_include())
  environment  = local.path_parts[0]
  service      = local.path_parts[1]
  
  # Common tags
  common_tags = {
    Project     = "IntegracionDeAplicaciones2"
    Environment = local.environment
    Service     = local.service
    ManagedBy   = "Terragrunt"
  }
}

# Configure remote state
remote_state {
  backend = "s3"
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
  config = {
    bucket  = "integracionbucket"
    key     = "terraform-state/${local.environment}/${local.service}/terraform.tfstate"
    region  = "us-east-1"
    encrypt = true
  }
}

# Generate provider configuration
generate "provider" {
  path = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents = <<EOF
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
  region = var.aws_region
  
  default_tags {
    tags = var.tags
  }
}
EOF
}
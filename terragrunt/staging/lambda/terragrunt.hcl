include "root" {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../infrastructure/terraform/lambda"
}

inputs = {
  environment = "staging"
  lambda_function_name = "helloWorld"
  aws_region  = "us-east-1"
  
  # These will be provided by environment variables in CI/CD
  lambda_s3_bucket = get_env("TF_VAR_lambda_s3_bucket", "")
  lambda_s3_key    = get_env("TF_VAR_lambda_s3_key", "")
  
  tags = {
    Project     = "IntegracionDeAplicaciones2"
    Environment = "staging"
    Service     = "lambda"
    ManagedBy   = "Terragrunt"
  }
}
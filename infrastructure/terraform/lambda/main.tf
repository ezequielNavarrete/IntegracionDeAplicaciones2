# Use existing LabRole instead of creating new IAM resources
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}


resource "aws_lambda_function" "lambda" {
  function_name = var.lambda_function_name
  role          = data.aws_iam_role.lab_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  timeout       = 30
  memory_size   = 128

  s3_bucket = var.lambda_s3_bucket
  s3_key    = var.lambda_s3_key

  environment {
    variables = {
      ENVIRONMENT = var.environment
    }
  }

  tags = merge(var.tags, {
    Name        = "${var.lambda_function_name}"
    Environment = var.environment
  })
}

# Lambda function URL (for easy HTTP access)
resource "aws_lambda_function_url" "lambda_url" {
  function_name      = aws_lambda_function.lambda.function_name
  authorization_type = "NONE"

  cors {
    allow_credentials = false
    allow_origins     = ["*"]
    allow_methods     = ["GET", "POST"]
    allow_headers     = ["date", "keep-alive"]
    expose_headers    = ["date", "keep-alive"]
    max_age          = 86400
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${var.lambda_function_name}"
  retention_in_days = 14

  tags = merge(var.tags, {
    Name        = "${var.lambda_function_name}-logs"
    Environment = var.environment
  })
}
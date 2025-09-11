# IAM role for Lambda function
resource "aws_iam_role" "lambda_role" {
  name = "hello-world-lambda-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name        = "hello-world-lambda-role-${var.environment}"
    Environment = var.environment
  })
}

# Basic Lambda execution policy
resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Policy for S3 access (for SDK example)
resource "aws_iam_policy" "lambda_s3_policy" {
  name        = "hello-world-lambda-s3-policy-${var.environment}"
  description = "Policy for Lambda to access S3"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:ListAllMyBuckets",
          "s3:GetBucketLocation"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(var.tags, {
    Name        = "hello-world-lambda-s3-policy-${var.environment}"
    Environment = var.environment
  })
}

resource "aws_iam_role_policy_attachment" "lambda_s3_policy_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_s3_policy.arn
}

# Lambda function
resource "aws_lambda_function" "hello_world" {
  function_name = "hello-world-lambda-${var.environment}"
  role          = aws_iam_role.lambda_role.arn
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
    Name        = "hello-world-lambda-${var.environment}"
    Environment = var.environment
  })
}

# Lambda function URL (for easy HTTP access)
resource "aws_lambda_function_url" "hello_world_url" {
  function_name      = aws_lambda_function.hello_world.function_name
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
  name              = "/aws/lambda/hello-world-lambda-${var.environment}"
  retention_in_days = 14

  tags = merge(var.tags, {
    Name        = "hello-world-lambda-logs-${var.environment}"
    Environment = var.environment
  })
}
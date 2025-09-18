.PHONY: build test clean deploy-dev deploy-staging deploy-prod help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=bootstrap
BINARY_UNIX=$(BINARY_NAME)_linux_amd64
LAMBDA_ZIP=lambda-deployment.zip

# AWS parameters
AWS_REGION ?= us-east-1
S3_BUCKET ?= $(S3_DEPLOYMENT_BUCKET)
LAMBDA_FUNCTION_NAME ?= helloWorld

help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Targets:'
	@egrep '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Download Go dependencies
	$(GOMOD) download
	$(GOMOD) tidy

test: ## Run tests
	$(GOTEST) -v ./...

build: deps ## Build the Lambda function
	cd src/lambda && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_NAME) main.go

package: build ## Create deployment package
	cd src/lambda && \
	zip $(LAMBDA_ZIP) $(BINARY_NAME)

clean: ## Clean build artifacts
	$(GOCLEAN)
	cd src/lambda && rm -f $(BINARY_NAME) $(LAMBDA_ZIP)

upload: package ## Upload package to S3
	@if [ -z "$(S3_BUCKET)" ]; then \
		echo "Error: S3_BUCKET not set. Use: make upload S3_BUCKET=your-bucket-name"; \
		exit 1; \
	fi
	cd src/lambda && \
	aws s3 cp $(LAMBDA_ZIP) s3://$(S3_BUCKET)/lambda-deployments/$(LAMBDA_FUNCTION_NAME)-latest.zip

deploy-dev: ## Deploy to dev environment
	@if [ -z "$(S3_BUCKET)" ]; then \
		echo "Error: S3_BUCKET not set. Use: make deploy-dev S3_BUCKET=your-bucket-name"; \
		exit 1; \
	fi
	cd terragrunt/dev/lambda && \
	export TF_VAR_lambda_s3_bucket=$(S3_BUCKET) && \
	export TF_VAR_lambda_s3_key="lambda-deployments/$(LAMBDA_FUNCTION_NAME)-latest.zip" && \
	export TF_VAR_environment="dev" && \
	export TF_VAR_lambda_function_name="$(LAMBDA_FUNCTION_NAME)" && \
	terragrunt apply -auto-approve

deploy-staging: ## Deploy to staging environment
	@if [ -z "$(S3_BUCKET)" ]; then \
		echo "Error: S3_BUCKET not set. Use: make deploy-staging S3_BUCKET=your-bucket-name"; \
		exit 1; \
	fi
	cd terragrunt/staging/lambda && \
	export TF_VAR_lambda_s3_bucket=$(S3_BUCKET) && \
	export TF_VAR_lambda_s3_key="lambda-deployments/$(LAMBDA_FUNCTION_NAME)-latest.zip" && \
	export TF_VAR_environment="staging" && \
	export TF_VAR_lambda_function_name="$(LAMBDA_FUNCTION_NAME)" && \
	terragrunt apply -auto-approve

deploy-prod: ## Deploy to prod environment
	@if [ -z "$(S3_BUCKET)" ]; then \
		echo "Error: S3_BUCKET not set. Use: make deploy-prod S3_BUCKET=your-bucket-name"; \
		exit 1; \
	fi
	cd terragrunt/prod/lambda && \
	export TF_VAR_lambda_s3_bucket=$(S3_BUCKET) && \
	export TF_VAR_lambda_s3_key="lambda-deployments/$(LAMBDA_FUNCTION_NAME)-latest.zip" && \
	export TF_VAR_environment="prod" && \
	export TF_VAR_lambda_function_name="$(LAMBDA_FUNCTION_NAME)" && \
	terragrunt apply -auto-approve

destroy-dev: ## Destroy dev environment
	cd terragrunt/dev/lambda && terragrunt destroy -auto-approve

destroy-staging: ## Destroy staging environment
	cd terragrunt/staging/lambda && terragrunt destroy -auto-approve

destroy-prod: ## Destroy prod environment
	cd terragrunt/prod/lambda && terragrunt destroy -auto-approve

test-lambda: ## Test the deployed Lambda function
	@ENV=$${ENV:-dev}; \
	FUNCTION_NAME=$(LAMBDA_FUNCTION_NAME); \
	echo "Testing Lambda function: $$FUNCTION_NAME"; \
	aws lambda invoke \
		--function-name "$$FUNCTION_NAME" \
		--payload '{"httpMethod": "GET", "path": "/", "requestContext": {"requestId": "test-request", "requestTime": "2024-01-01T00:00:00Z"}}' \
		response.json && \
	cat response.json && \
	rm response.json

local-test: ## Run Lambda function locally (requires SAM CLI)
	cd src/lambda && \
	sam local invoke -e ../../test-event.json

init-tf: ## Initialize Terraform in all environments
	cd terragrunt/dev/lambda && terragrunt init
	cd terragrunt/staging/lambda && terragrunt init
	cd terragrunt/prod/lambda && terragrunt init
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Response represents the Lambda response structure
type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// HelloWorldResponse represents the response body structure
type HelloWorldResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	log.Printf("Processing request: %s", request.RequestContext.RequestID)

	// Initialize AWS SDK v2
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error loading AWS config: %v", err)
		return Response{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to initialize AWS SDK"}`,
		}, err
	}

	// Example of using AWS SDK v2 - create S3 client
	s3Client := s3.NewFromConfig(cfg)
	
	// List buckets as an example of SDK usage
	bucketsResult, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	var bucketCount int
	if err != nil {
		log.Printf("Warning: Could not list S3 buckets: %v", err)
		bucketCount = -1 // Indicate error but don't fail the request
	} else {
		bucketCount = len(bucketsResult.Buckets)
	}

	// Create response
	responseBody := HelloWorldResponse{
		Message:   fmt.Sprintf("Hello World from AWS Lambda! You have %d S3 buckets.", bucketCount),
		RequestID: request.RequestContext.RequestID,
		Timestamp: request.RequestContext.RequestTime,
	}

	responseJSON, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return Response{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to marshal response"}`,
		}, err
	}

	return Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}

func main() {
	lambda.Start(handler)
}
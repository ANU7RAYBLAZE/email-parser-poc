package dynamodb

import (
	"context"
	"email-parser-poc/internal/domain/entities"
	"email-parser-poc/internal/ports/outgoing"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DB struct {
	DynamoClient *dynamodb.Client
}

func NewDynamoDbInstance(localstackEndpoint string) (outgoing.DbService, error) {
	// Load config with LocalStack endpoint and static credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test-key", "test-secret", "test-session"),
		),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolver(
			aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           localstackEndpoint,
					SigningRegion: "us-east-1",
				}, nil
			}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	return &DB{DynamoClient: client}, nil
}

func (d *DB) UploadHeaders(ctx context.Context, emails *entities.EmailList) error {
	if d.DynamoClient == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	timestamp := time.Now().Format(time.RFC3339)
	for _, email := range emails.Emails {
		var writeRequests []types.WriteRequest

		// Prepare items
		for headerName, headerValue := range email.Headers {
			writeRequests = append(writeRequests, types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"email_id":     &types.AttributeValueMemberS{Value: email.ID},
						"header_name":  &types.AttributeValueMemberS{Value: headerName},
						"header_value": &types.AttributeValueMemberS{Value: headerValue},
						"timestamp":    &types.AttributeValueMemberS{Value: timestamp},
					},
				},
			})
		}

		// Batch write in chunks of 25
		for i := 0; i < len(writeRequests); i += 25 {
			end := i + 25
			if end > len(writeRequests) {
				end = len(writeRequests)
			}
			batch := writeRequests[i:end]

			_, err := d.DynamoClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{
					"gmail-headers": batch,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to batch write headers: %w", err)
			}
		}

		fmt.Printf("✅ Uploaded %d headers for email %s\n", len(email.Headers), email.ID)
	}
	return nil
}

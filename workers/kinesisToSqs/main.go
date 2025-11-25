package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	ctx := context.Background()

	// Load AWS config
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		localstackURL := getEnv("LOCALSTACK_URL", "http://localhost:4566")
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           localstackURL,
			SigningRegion: "us-east-1",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}, nil
		})),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	kinesisClient := kinesis.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	streamName := getEnv("KINESIS_STREAM_NAME", "intervention-events")
	queueURL := getEnv("SQS_QUEUE_URL", "http://localhost:4566/000000000000/intervention-events")

	log.Printf("Starting Kinesis to SQS forwarder...")
	log.Printf("Kinesis stream: %s", streamName)
	log.Printf("SQS queue: %s", queueURL)

	// Get shard iterator
	shardIterator, err := getShardIterator(ctx, kinesisClient, streamName)
	if err != nil {
		log.Fatalf("failed to get shard iterator: %v", err)
	}

	// Poll Kinesis and forward to SQS
	for {
		records, nextIterator, err := getRecords(ctx, kinesisClient, shardIterator)
		if err != nil {
			log.Printf("Error getting records: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, record := range records {
			if err := forwardToSQS(ctx, sqsClient, queueURL, record); err != nil {
				log.Printf("Error forwarding to SQS: %v", err)
			} else {
				log.Printf("Forwarded event to SQS: %s", string(record.Data))
			}
		}

		shardIterator = nextIterator
		time.Sleep(2 * time.Second)
	}
}

func getShardIterator(ctx context.Context, client *kinesis.Client, streamName string) (*string, error) {
	describeResp, err := client.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		return nil, err
	}

	if len(describeResp.StreamDescription.Shards) == 0 {
		return nil, nil
	}

	shardId := describeResp.StreamDescription.Shards[0].ShardId

	resp, err := client.GetShardIterator(ctx, &kinesis.GetShardIteratorInput{
		StreamName:        aws.String(streamName),
		ShardId:           shardId,
		ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
	})
	if err != nil {
		return nil, err
	}

	return resp.ShardIterator, nil
}

func getRecords(ctx context.Context, client *kinesis.Client, shardIterator *string) ([]types.Record, *string, error) {
	resp, err := client.GetRecords(ctx, &kinesis.GetRecordsInput{
		ShardIterator: shardIterator,
		Limit:         aws.Int32(10),
	})
	if err != nil {
		return nil, nil, err
	}

	return resp.Records, resp.NextShardIterator, nil
}

func forwardToSQS(ctx context.Context, client *sqs.Client, queueURL string, record types.Record) error {
	_, err := client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(record.Data)),
	})
	return err
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"log"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:4566"}, nil
			})),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	kc := kinesis.NewFromConfig(cfg)

	_, err = kc.PutRecord(context.TODO(), &kinesis.PutRecordInput{
		Data:         []byte(`{"event":"patient_created","data":{"id":"123","name":"John Doe"}}`),
		PartitionKey: aws.String("1"),
		StreamName:   aws.String("cdc-stream"),
	})

	if err != nil {
		log.Fatalf("failed to put record, %v", err)
	}

	fmt.Println("Successfully published CDC event to Kinesis")
}

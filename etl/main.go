package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context, kinesisEvent events.KinesisEvent) error {
	for _, record := range kinesisEvent.Records {
		kinesisRecord := record.Kinesis
		dataBytes := kinesisRecord.Data
		dataString := string(dataBytes)

		fmt.Printf("Received record: %s\n", dataString)

		// In a real application, you would process the data and
		// send it to an SQS queue
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}

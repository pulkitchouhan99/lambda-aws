#!/bin/bash

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1

echo "Waiting for LocalStack to be ready..."
until aws --endpoint-url=http://localhost:4566 s3 ls > /dev/null 2>&1; do
  sleep 1
done

echo "Creating Kinesis streams..."
aws --endpoint-url=http://localhost:4566 kinesis create-stream --stream-name cdc-stream --shard-count 1
aws --endpoint-url=http://localhost:4566 kinesis create-stream --stream-name intervention-events --shard-count 1

echo "Creating SQS queues..."
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name patient-dlq
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name patient --attributes '{ "RedrivePolicy": "{\"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:000000000000:patient-dlq\",\"maxReceiveCount\":\"5\"}" }'

aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name screening-dlq
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name screening --attributes '{ "RedrivePolicy": "{\"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:000000000000:screening-dlq\",\"maxReceiveCount\":\"5\"}" }'

aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name intervention-dlq
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name intervention-events --attributes '{ "RedrivePolicy": "{\"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:000000000000:intervention-dlq\",\"maxReceiveCount\":\"5\"}" }'

echo "Creating S3 bucket..."
aws --endpoint-url=http://localhost:4566 s3 mb s3://audit-archive

echo "LocalStack initialized successfully!"
echo "Kinesis streams: cdc-stream, intervention-events"
echo "SQS queues: intervention-events, patient, screening"


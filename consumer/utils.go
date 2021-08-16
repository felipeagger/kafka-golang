package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func init() {
	return

	if LOCALSTACK == "" {
		LOCALSTACK = "http://localhost:4576"
	}

	// config := &aws.Config{
	// 	Region: aws.String("sa-east-1"),
	// 	//Credentials: credentials.NewSharedCredentials("", "default"),
	// 	Endpoint: aws.String(LOCALSTACK),
	// }

	// sess := session.Must(session.NewSession(config))
	// sqsClient = sqs.New(sess)

	client, err := NewSQSClient(LOCALSTACK, "us-east-1")
	if err != nil {
		panic(err)
	}

	sqsClient = client

	createQueue("sqs-events-dlq", "15", "86400")
	fmt.Println("AWS Session initialized...")
}

func NewSQSClient(awsEndpoint, awsRegion string) (*sqs.Client, error) {

	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if awsEndpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsEndpoint,
				SigningRegion: awsRegion,
			}, nil
		}

		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithEndpointResolver(customResolver),
	)

	if err != nil {
		return nil, err
	}

	client := sqs.NewFromConfig(awsCfg)

	return client, nil
}

func createQueue(name, delay, retentionPeriod string) {

	queue, err := sqsClient.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})

	if err != nil {
		result, err := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
			QueueName: aws.String(name),
			Attributes: map[string]string{
				"DelaySeconds":           delay,
				"MessageRetentionPeriod": retentionPeriod,
			},
		})

		if err != nil {
			panic(err)
		}

		DEADQUEUE = *result.QueueUrl
	} else {
		DEADQUEUE = *queue.QueueUrl
	}
}

func sendMessage(data string) {

	_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: aws.String(data),
		QueueUrl:    aws.String(DEADQUEUE),
	})

	checkError("Send Msg", err)
}

func deleteMessage(msgHandle *string) {

	_, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(DEADQUEUE),
		ReceiptHandle: msgHandle,
	})

	checkError("Delete Msg", err)
}

func checkError(_type string, err error) {
	if err != nil {
		fmt.Printf("\nError on %s : %v", _type, err)
	}
}

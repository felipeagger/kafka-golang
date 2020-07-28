package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func init() {

	config := &aws.Config{
		Region:      aws.String("sa-east-1"),
		Credentials: credentials.NewSharedCredentials("", "default"),
		Endpoint:    aws.String("http://localhost:4576"),
	}

	sess := session.Must(session.NewSession(config))
	sqsClient = sqs.New(sess)

	createQueue("sqs-events-dlq")
	fmt.Println("AWS Session initialized...")
}

func createQueue(name string) {

	queue, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})

	if err != nil {
		result, _ := sqsClient.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(name),
			Attributes: map[string]*string{
				"DelaySeconds":           aws.String("15"),
				"MessageRetentionPeriod": aws.String("86400"),
			},
		})

		deadQueueName = *result.QueueUrl
	} else {
		deadQueueName = *queue.QueueUrl
	}
}

func sendMessage(data string) {

	_, err := sqsClient.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(data),
		QueueUrl:    aws.String(deadQueueName),
	})

	checkError("Send Msg", err)
}

func deleteMessage(msgHandle *string) {

	_, err := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(deadQueueName),
		ReceiptHandle: msgHandle,
	})

	checkError("Delete Msg", err)
}

func checkError(_type string, err error) {
	if err != nil {
		fmt.Printf("\nError on %s : %v", _type, err)
	}
}

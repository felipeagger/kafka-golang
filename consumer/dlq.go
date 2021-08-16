package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var waitGrpDlq sync.WaitGroup

func consumeDLQ(chanend chan bool) {

	ticker := time.Tick(time.Second * 10)
	run := true
	for run {
		select {
		case <-chanend:
			fmt.Println("Terminating DLQ")
			run = false

		case <-ticker:

			output, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
				QueueUrl: aws.String(DEADQUEUE),
				//MaxNumberOfMessages: aws.Int32(10),
				//VisibilityTimeout:   aws.Int32(30),
			})

			if err != nil {
				continue
			}

			for _, msg := range output.Messages {
				waitGrpDlq.Add(1)
				go processDlqMsg(msg)
			}

			waitGrpDlq.Wait()

		}
	}

}

func processDlqMsg(msg types.Message) {
	defer waitGrpDlq.Done()
	var err error

	//process

	if uint64(rand.Intn(10)) > 7 {
		err = errors.New("Falhou DLQ")
	}

	if err != nil {
		fmt.Printf("\nError on process msg DLQ: %s", *msg.Body)
		return
	}

	fmt.Printf("\nSucess processed msg DLQ: %s\n", *msg.Body)
	deleteMessage(msg.ReceiptHandle)
}

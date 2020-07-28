package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go/service/sqs"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var (
	topic         string = os.Getenv("TOPIC")
	group         string = os.Getenv("GROUP")
	brokerIP      string = os.Getenv("BROKER_SRV")
	brokerPort    string = os.Getenv("BROKER_PORT")
	deadQueueName string
	waitGrp       sync.WaitGroup
	consumer      *kafka.Consumer
	sqsClient     *sqs.SQS
)

func init() {
	config := kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%s:%s", brokerIP, brokerPort),
		"group.id":          group,
		"auto.offset.reset": "earliest",
		//"isolation.level":    "read_committed",
		//"enable.auto.commit": false,
		"session.timeout.ms": 10000,
	}

	var err error
	consumer, err = kafka.NewConsumer(&config)

	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Printf("Initializing Consumer...\nConsumerGroup: %v \nTopic: %v\n",
		group, topic)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	chanend := make(chan bool)

	go consumeDLQ(chanend)
	consume(sigchan, chanend)
}

func consume(sigchan chan os.Signal, chanend chan bool) {

	consumer.SubscribeTopics([]string{topic}, nil)

	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("\nCaught signal %v: terminating\n", sig)
			chanend <- true
			run = false

		default:
			ev := consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch event := ev.(type) {

			case *kafka.Message:

				waitGrp.Add(1)
				go processMsg(event)

			case kafka.Error:
				fmt.Fprintf(os.Stderr, "\n%% Error: %v: %v\n", event.Code(), event)
				if event.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("\nIgnored %v\n", event)
			}

		}
	}

	waitGrp.Wait()
	consumer.Close()
}

func processMsg(event *kafka.Message) {
	defer waitGrp.Done()

	//event.Value, event.Key, event.Headers, event.TopicPartition
	data := string(event.Value)

	var err error

	if uint64(rand.Intn(10)) > 7 {
		err = errors.New("Falhou")
	}

	if err != nil {
		fmt.Printf("Sending Msg to DLQ %s: %v\n", event.TopicPartition.Offset, data)
		sendMessage(data)
		return
	}

	fmt.Printf("Msg %v on OffSet: %s\n", data, event.TopicPartition.Offset)
}

package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var (
	TOPIC      string = os.Getenv("TOPIC")
	GROUP      string = os.Getenv("GROUP")
	BROKERS    string = os.Getenv("BROKERS")
	LOCALSTACK string = os.Getenv("LOCALSTACK")
	DEADQUEUE  string // Automatic set on utils
	waitGrp    sync.WaitGroup
	consumer   *kafka.Consumer
	sqsClient  *sqs.Client
)

func init() {
	if BROKERS == "" {
		panic("Missing environment variable: BROKERS")
	}

	if TOPIC == "" {
		panic("Missing environment variable: TOPIC")
	}

	config := kafka.ConfigMap{
		"bootstrap.servers": BROKERS,
		"group.id":          GROUP,
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
	fmt.Printf("Initializing Consumer...\nConsumerGroup: %v \nTopic: %v\n", GROUP, TOPIC)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	chanend := make(chan bool)

	//go consumeDLQ(chanend)
	consume(sigchan, chanend)
}

func consume(sigchan chan os.Signal, chanend chan bool) {

	consumer.SubscribeTopics([]string{TOPIC}, nil)

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

	fmt.Printf("Timestamp: %v\n Offset: %v\n eventHeaders: %v\n eventKey: %v\n", event.Timestamp, event.TopicPartition.Offset, event.Headers, event.Key)

	data := string(event.Value)

	var err error
	//if uint64(rand.Intn(10)) > 7 {
	//	err = errors.New("Falhou")
	//}

	if err != nil {
		fmt.Printf("Sending Msg to DLQ %s: %v\n", event.TopicPartition.Offset, data)
		sendMessage(data)
		return
	}

	fmt.Printf("Msg %v on OffSet: %s\n", data, event.TopicPartition.Offset)
}

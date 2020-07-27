package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	topic      string = os.Getenv("TOPIC")
	group      string = os.Getenv("GROUP")
	brokerIP   string = os.Getenv("BROKER_SRV")
	brokerPort string = os.Getenv("BROKER_PORT")
	consumer   *kafka.Consumer
	waitGrp    sync.WaitGroup
)

func init() {
	config := kafka.ConfigMap{
		"bootstrap.servers":  fmt.Sprintf("%s:%s", brokerIP, brokerPort),
		"group.id":           group,
		"auto.offset.reset":  "earliest",
		"session.timeout.ms": 10000,
	}

	var err error
	fmt.Println("Initializing consumer...")
	consumer, err = kafka.NewConsumer(&config)

	if err != nil {
		panic(err)
	}
}

func main() {

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	consume(sigchan)
}

func consume(sigchan chan os.Signal) {

	consumer.SubscribeTopics([]string{topic}, nil)

	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false

		default:
			ev := consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch event := ev.(type) {

			case *kafka.Message:

				waitGrp.Add(1)
				go processMsg(event.Value, event.Key, event.Headers, event.TopicPartition)

			case kafka.Error:
				fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", event.Code(), event)
				if event.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", event)
			}

		}
	}

	waitGrp.Wait()
	consumer.Close()
}

func processMsg(value, key []byte, headers []kafka.Header, topic kafka.TopicPartition) {
	defer waitGrp.Done()

	data := string(value)

	//if generate an error here?

	fmt.Printf("Msg %v on topic %s\n", data, topic)
}

package main

import (
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	TOPIC    string = os.Getenv("TOPIC")
	BROKERS  string = os.Getenv("BROKERS")
	producer *kafka.Producer
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
		//"enable.idempotence": true,
		//"acks":               "all",
		"retries": 10,
	}

	var err error
	producer, err = kafka.NewProducer(&config)

	if err != nil {
		panic(err)
	}
}

func main() {

	termChan := make(chan bool, 1)
	doneChan := make(chan bool)

	go listenEvents(termChan, doneChan, producer)

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf(`{"msg": "kafka with Golang - %v"}`, i)

		sendMessage(fmt.Sprintf("msg%v", i), msg)
	}

	closeProducer(producer, termChan, doneChan)

}

func sendMessage(key, data string) {

	var headers []kafka.Header
	headers = append(headers, kafka.Header{
		Key:   "origin",
		Value: []byte("producer"),
	})

	message := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &TOPIC,
			Partition: kafka.PartitionAny,
		},
		Headers: headers,
		Key:     []byte(key),
		Value:   []byte(data),
	}

	if err := producer.Produce(&message, nil); err != nil {
		fmt.Printf("Error on producing message! %v", err.Error())
	}

}

func closeProducer(producer *kafka.Producer, termChan, doneChan chan bool) {
	// Flush the Producer queue
	timeOut := 10000
	if count := producer.Flush(timeOut); count > 0 {
		fmt.Printf("\nFailed to flush messages. %d message(s) remain\n", count)
	} else {
		fmt.Println("All messages flushed from the queue!")
	}

	// Stop listening to events and close the producer
	termChan <- true
	<-doneChan
	producer.Close()
}

// Handle any events that we get
func listenEvents(termChan, doneChan chan bool, producer *kafka.Producer) {
	doTerm := false
	for !doTerm {
		select {
		case ev := <-producer.Events():
			switch ev.(type) {
			case *kafka.Message:
				km := ev.(*kafka.Message)

				if km.TopicPartition.Error != nil {

					fmt.Printf("Failed to send message '%v' to topic '%v'\n\tErr: %v",
						string(km.Value),
						string(*km.TopicPartition.Topic),
						km.TopicPartition.Error)

				} else {

					fmt.Printf("Message '%v' delivered to topic '%v' (partition %d at offset %d)\n",
						string(km.Value),
						string(*km.TopicPartition.Topic),
						km.TopicPartition.Partition,
						km.TopicPartition.Offset)
				}

			case kafka.Error:
				em := ev.(kafka.Error)
				fmt.Printf("Error:\n\t%v\n", em)
			}

		case <-termChan:
			doTerm = true

		}
	}
	close(doneChan)
}

package main

import (
	"fmt"
	"os"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var (
	topic      string = os.Getenv("TOPIC")
	brokerIP   string = os.Getenv("BROKER_SRV")
	brokerPort string = os.Getenv("BROKER_PORT")
	producer   *kafka.Producer
)

func init() {
	config := kafka.ConfigMap{
		"bootstrap.servers":  fmt.Sprintf("%s:%s", brokerIP, brokerPort),
		"enable.idempotence": true,
		"acks":               "all",
		"retries":            10,
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
		sendMessage(fmt.Sprintf("kafka with Golang - %v", i))
	}

	closeProducer(producer, termChan, doneChan)

}

func sendMessage(data string) {

	message := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(data),
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

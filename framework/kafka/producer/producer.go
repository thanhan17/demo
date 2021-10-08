package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

var (
	kafkaBrokers = "localhost:9093,localhost:9094,localhost:9095"
	brokerList   = strings.Split(kafkaBrokers, ",")
	KafkaTopic   = "mytopic"
	enqueued     int
)

func main() {

	producer, err := setupProducer()
	if err != nil {
		panic(err)
	} else {
		log.Println("Kafka AsyncProducer up and running!")
	}

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	produceMessages(producer, signals)

	log.Printf("Kafka AsyncProducer finished with %d messages produced.", enqueued)
}

// setupProducer will create a AsyncProducer and returns it
func setupProducer() (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	return sarama.NewAsyncProducer(brokerList, config)
}

// produceMessages will send 'testing 123' to KafkaTopic each second, until receive a os signal to stop e.g. control + c
// by the user in terminal
func produceMessages(producer sarama.AsyncProducer, signals chan os.Signal) {
	var count int = 0
	for {
		count += 1
		time.Sleep(time.Second)
		mess := "message " + strconv.Itoa(count)
		message := &sarama.ProducerMessage{Topic: KafkaTopic, Value: sarama.StringEncoder(mess), Partition: int32(count % 4)}
		select {
		case producer.Input() <- message:
			enqueued++
			log.Println("New Message produced: " + mess)
		case <-signals:
			producer.AsyncClose() // Trigger a shutdown of the producer.
			return
		}
	}
}

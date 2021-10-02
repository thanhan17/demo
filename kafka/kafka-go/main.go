package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

const (
	topic          = "mytopic"
	broker1Address = "localhost:9093"
	broker2Address = "localhost:9094"
	broker3Address = "localhost:9095"
	partition      = 4
)

func produce(ctx context.Context) {
	// Counter
	i := 0

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker1Address, broker2Address, broker3Address},
		Topic:   topic,
	})

	for {
		err := w.WriteMessages(ctx, kafka.Message{
			Key:    []byte(strconv.Itoa(i % 4)),
			Value:  []byte("message:" + strconv.Itoa(i)),
			Offset: int64(i),
		})
		if err != nil {
			panic("Could not write message " + err.Error())
		}

		// log a confirmation once the message is written
		fmt.Println("writes:", i)
		i++
		// sleep for a second
		time.Sleep(time.Millisecond * 100)
	}
}

func consume(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker1Address, broker2Address, broker3Address},
		GroupID:  "mygroup",
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		fmt.Println("received: ", string(msg.Value))
	}
}

func main() {
	ctx := context.Background()
	produce(ctx)
	//consume(ctx)
}

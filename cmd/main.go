package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"
)

func main() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "micro-kafka:9092",
		"acks":              "all",
	})

	if err != nil {
		log.Printf("Failed to create producer: %s", err)
		os.Exit(1)
	}

	// Go-routine to handle events
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Produced event to topic %s: key = %-10s value = %s\n", *ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

	// users := [...]string{"eabara", "jsmith", "sgarcia", "jbernard", "htanaka", "awalther"}
	// items := [...]string{"book", "alarm clock", "t-shirts", "gift card", "batteries"}
	// topic := "purchases"

	// for n := 0; n < 10; n++ {
	// 	key := users[rand.Intn(len(users))]
	// 	data := items[rand.Intn(len(items))]
	// 	p.Produce(&kafka.Message{
	// 		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
	// 		Key:            []byte(key),
	// 		Value:          []byte(data),
	// 	}, nil)
	// }

	mux := gin.Default()

	httpServer := &http.Server{
		Addr:           ":8090",
		Handler:        mux,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	log.Println("hello world")
	// })

	mux.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	httpServer.ListenAndServe()

	// Wait for all messages to be delivered
	p.Flush(15 * 1000)
	p.Close()
}

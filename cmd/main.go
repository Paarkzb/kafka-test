package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"
)

type SomeData struct {
	User string `json:"user"`
	Item string `json:"item"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status string `json:"status"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	// logrus.Error(message)
	c.AbortWithStatusJSON(statusCode, errorResponse{Message: message})
}

func main() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("BOOTSTRAP_SERVERS"),
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

	pool, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close(context.Background())

	mux := gin.Default()

	httpServer := &http.Server{
		Addr:           ":8090",
		Handler:        mux,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	mux.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	mux.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Home",
		})
	})
	mux.POST("/", func(c *gin.Context) {

		var input SomeData
		if err := c.BindJSON(&input); err != nil {
			// logrus.Println("HERE")
			newErrorResponse(c, http.StatusBadRequest, "binding data "+err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})

	httpServer.ListenAndServe()

	// Wait for all messages to be delivered
	p.Flush(15 * 1000)
	p.Close()
}

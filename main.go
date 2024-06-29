// main.go
package main

import (
	"context"
	"log"

	"github.com/pmas98/go-auth-service/routes"
	"github.com/pmas98/go-auth-service/utils"
)

func main() {
	r := routes.SetupRouter()
	consumerGroup, err := utils.InitKafkaConsumerGroup("auth-service-consumer-group")
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer group: %v", err)
	}
	defer consumerGroup.Close()

	go func() {
		err := consumerGroup.Consume(context.Background(), []string{"token_verification_requests"}, &utils.ConsumerGroupHandler{})
		if err != nil {
			log.Fatalf("Error consuming messages from Kafka: %v", err)
		}
	}()

	r.Run(":8082")
}

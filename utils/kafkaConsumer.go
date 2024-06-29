package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/pmas98/go-auth-service/models"
)

func InitKafkaConsumerGroup(groupID string) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(kafkaBrokers, groupID, config)
	if err != nil {
		return nil, err
	}
	return consumerGroup, nil
}

func ConsumeMessagesFromKafka(topic string, groupID string) error {
	consumerGroup, err := InitKafkaConsumerGroup(groupID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	handler := &ConsumerGroupHandler{}

	for {
		err := consumerGroup.Consume(ctx, []string{topic}, handler)
		if err != nil {
			log.Printf("Error from consumer: %v", err)
			return err
		}
	}
}

// ConsumerGroupHandler represents a Sarama consumer group consumer
type ConsumerGroupHandler struct{}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message received: value = %s, timestamp = %v, topic = %s\n", string(msg.Value), msg.Timestamp, msg.Topic)

		var tokenStruct struct {
			Token string `json:"token"`
		}

		// Unmarshal msg.Value into the struct
		err := json.Unmarshal(msg.Value, &tokenStruct)
		if err != nil {
			log.Printf("Error unmarshaling token JSON: %v", err)
			continue
		}

		token := tokenStruct.Token

		valid, userID, name, email := verifyToken(token)

		response := models.TokenVerificationResponse{
			Valid:  valid,
			UserID: userID,
			Email:  email,
			Name:   name,
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			continue
		}

		InitKafkaProducer()
		SendMessageJSONToKafka("token_verification_responses", responseJSON, strconv.Itoa(userID))

		session.MarkMessage(msg, "")
	}
	return nil
}

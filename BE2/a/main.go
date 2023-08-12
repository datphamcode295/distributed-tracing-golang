package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp" // RabbitMQ package
)

type QueueMessage struct {
	TraceID       string          `json:"trace_id"`
	ParentTraceID string          `json:"parent_trace_id"`
	Email         string          `json:"email"`
	Ctx           context.Context `json:"ctx"`
}

func main() {
	// Establish a connection to RabbitMQ server
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Your QueueMessage object
	message := QueueMessage{
		TraceID:       "123456",
		ParentTraceID: "789012",
		Email:         "example@example.com",
		Ctx:           context.Background(),
	}

	// Convert the message to JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Failed to marshal message to JSON: %v", err)
	}

	// Publish the message to a queue
	err = ch.Publish(
		"",           // Exchange
		"your_queue", // Queue name
		false,        // Mandatory
		false,        // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBytes,
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	fmt.Println("Message sent successfully")
}

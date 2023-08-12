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

	// Declare the queue
	q, err := ch.QueueDeclare(
		"your_queue", // Queue name
		false,        // Durable
		false,        // Delete when unused
		false,        // Exclusive
		false,        // No-wait
		nil,          // Arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name, // Queue name
		"",     // Consumer
		true,   // Auto-acknowledge
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Arguments
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	// Process incoming messages
	for msg := range msgs {
		var message QueueMessage
		err := json.Unmarshal(msg.Body, &message)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// Handle the received message
		fmt.Printf("Received message: %+v\n", message)
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/datphamcode295/distributed-tracing/tracing"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type QueueMessage struct {
	TraceID       string `json:"trace_id"`
	ParentTraceID string `json:"parent_trace_id"`
	Email         string `json:"email"`
}

func main() {
	tp, tpErr := tracing.JaegerTraceProvider()
	if tpErr != nil {
		log.Fatal(tpErr)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"email_queue", // queue name
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare a queue:", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,
	)
	if err != nil {
		log.Fatal("Failed to register a consumer:", err)
	}

	fmt.Println("Waiting for messages...")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			handleMessage(d)
		}
	}()

	<-forever
}

func handleMessage(d amqp.Delivery) {
	const spanID = "handleSentMailMessage"
	msg := string(d.Body)

	log.SetFormatter(&log.JSONFormatter{})
	l := log.WithFields(log.Fields{
		"message": msg,
		"span_id": spanID,
	})
	l.Info("Received sent mail message !")

	headers := d.Headers
	convertedHeaders := make(map[string]string)
	for k, v := range headers {
		convertedHeaders[k] = v.(string)
	}

	ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(convertedHeaders))
	ctx, span := otel.Tracer("sentEmail").Start(ctx, "sentEmail")
	defer span.End()

	// parse d to QueueMessage struct
	var message QueueMessage
	err := json.Unmarshal(d.Body, &message)
	if err != nil {
		l.Errorf("Failed to unmarshal message: %v", err)
		return
	}

	fmt.Printf("Sent message: %s\n", message)
}

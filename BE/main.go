package main

import (
	"fmt"
	"net/http"

	"github.com/datphamcode295/distributed-tracing/models"
	"github.com/datphamcode295/distributed-tracing/queue"
	"github.com/datphamcode295/distributed-tracing/tracing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	tp, tpErr := tracing.JaegerTraceProvider()
	if tpErr != nil {
		log.Fatal(tpErr)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	r := gin.Default()

	//gin OTEL instrumentation
	r.Use(otelgin.Middleware("api-service"))

	// Apply CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}                                                // Update with your frontend URL
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} // Allow the required methods
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "X-Trace-Id", "X-Parent-Id"}
	r.Use(cors.New(config))

	r.POST("/users/reset-password", resetPasswordHandler)

	fmt.Println("Server started on :3000")
	r.Run(":3000")
}

func resetPasswordHandler(c *gin.Context) {
	const spanID = "resetPasswordHandler"
	newCtx, span := otel.Tracer(spanID).Start(c.Request.Context(), spanID)
	defer span.End()

	// Update c ctx to newCtx
	c.Request = c.Request.WithContext(newCtx)

	parentID := c.GetHeader("X-Parent-Id")
	traceID := c.GetHeader("X-Trace-Id")

	log.SetFormatter(&log.JSONFormatter{})
	l := log.WithFields(log.Fields{
		"parent_id": parentID,
		"trace_id":  traceID,
		"span_id":   spanID,
	})

	// add log fields

	l.Info("Receive reset password request !")

	span.SetAttributes(attribute.String("X-Parent-Id", parentID))
	span.SetAttributes(attribute.String("X-Trace-Id", traceID))

	var request models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		l.Error("Error parsing request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	queueMessage := queue.QueueMessage{
		TraceID:       traceID,
		ParentTraceID: spanID,
		Email:         request.Email,
	}

	err := queue.EnqueueMessage(c.Request.Context(), queueMessage)
	if err != nil {
		l.Error("Error enqueue message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Password reset email sent",
	})
}

// CorsMiddleware handles CORS for the API
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Trace-Id, X-Span-Id")
		c.Writer.Header().Set("Content-Type", "application/json")

		if c.Request.Method == http.MethodOptions {
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		c.Next()
	}
}

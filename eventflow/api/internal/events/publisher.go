package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Function  string                 `json:"function"`
	Image     string                 `json:"image"`     // Function's container image
	Command   []string               `json:"command"`   // Function's command
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

type Publisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewPublisher(natsURL string) (*Publisher, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	// Ensure stream exists
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "EVENTFLOW",
		Subjects: []string{"eventflow.>"},
		MaxAge:   24 * time.Hour,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return &Publisher{nc: nc, js: js}, nil
}

func (p *Publisher) Publish(eventType, function string, payload map[string]interface{}) error {
	return p.PublishWithMetadata(eventType, function, "", nil, payload)
}

func (p *Publisher) PublishWithMetadata(eventType, function, image string, command []string, payload map[string]interface{}) error {
	event := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Function:  function,
		Image:     image,
		Command:   command,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = p.js.Publish("eventflow.events", data)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}

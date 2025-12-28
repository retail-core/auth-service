package queue

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/retail-core/auth-service/internal/logger"
	"go.uber.org/zap"
)

type Publisher struct {
	Ch *amqp.Channel
}

// NewPublisher sets up the RabbitMQ connection and channel once at startup.
func NewPublisher(rabbitURL string) *Publisher {
	var conn *amqp.Connection
	var err error

	for i := 0; i <= 10; i++ {
		conn, err = amqp.Dial(rabbitURL)

		if err == nil {
			break
		}
		
		logger.L().Error("Failed to connect to RabbitMQ, retrying...", zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(3 * time.Second)
	}

    if err != nil {
		logger.L().Fatal("Could not connect to RabbitMQ after several attempts", zap.Error(err))
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.L().Error("Failed to open RabbitMQ channel: %v", zap.Error(err))
	}
	logger.L().Info("✅ RabbitMQ publisher connected.")
	return &Publisher{Ch: ch}
}

// PublishNotification sends a JSON payload to the notifications queue.
func (p *Publisher) PublishNotification(ctx context.Context, routingKey string, msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.Ch	.PublishWithContext(
		ctx,
		"notifications", // exchange (use default direct)
		routingKey,      // queue name
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Publisher) PublishDomainEvent(ctx context.Context, routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.Ch.PublishWithContext(
		ctx,
		"domain.events", // ✅ publish to your domain exchange
		routingKey,      // e.g. "business_owner.created"
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Publisher) SetupDomainExchange() error {
	return p.Ch.ExchangeDeclare(
		"domain.events", // exchange name
		"topic",         // type (allows pattern routing)
		true,            // durable
		false,           // auto-delete
		false,           // internal
		false,           // no-wait
		nil,
	)
}

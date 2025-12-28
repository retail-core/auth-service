package consumer

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/retail-core/auth-service/internal/auth"
	"github.com/retail-core/auth-service/internal/dtos"
	"github.com/retail-core/auth-service/internal/logger"
	"go.uber.org/zap"
)

const (
	AuthExchangeName       	  = "domain.events"
	QueueName                 = "auth_service_queue"
	StaffDeletedRoutingKey   = "staff.deleted"
)

type AuthConsumer struct {
	Ch      *amqp.Channel
	Service auth.Service
}

func NewAuthConsumer(ch *amqp.Channel, service auth.Service) *AuthConsumer {

	return &AuthConsumer{
		Ch:      ch,
		Service: service,
	}
}

func (c *AuthConsumer) StartConsumption(ctx context.Context) error {

	log := logger.L()

	if err := c.Ch.ExchangeDeclare(AuthExchangeName, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	log.Info("📥 Declared DomainEvent exchange", zap.String("exchange Name", AuthExchangeName))

	q, err := c.Ch.QueueDeclare(QueueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Info("📥 Declared Account Service queue", zap.String("queue Name", q.Name))

	// 3. Bind Queue to Exchange for confirmation and rollback keys
	if err = c.Ch.QueueBind(q.Name, StaffDeletedRoutingKey, AuthExchangeName, false, nil); err != nil {
		return err
	}


	msgs, err := c.Ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		log.Info("✅ Queue bound and consuming",
			zap.String("queue", q.Name),
			zap.String("routing_keys", StaffDeletedRoutingKey),
		)
		for m := range msgs {
			c.handleMessage(ctx, m)
		}
	}()

	<-ctx.Done()
	return nil
}

func (c *AuthConsumer) handleMessage(ctx context.Context, m amqp.Delivery) {

	var err error

	switch m.RoutingKey {
	case StaffDeletedRoutingKey:
		var event dtos.StaffDeletedEvent

		if err := json.Unmarshal(m.Body, &event); err != nil {
			m.Ack(false)
			return
		}

		err = c.Service.DeleteUser(ctx, event.UserID)

	default:
		m.Ack(false)
		logger.L().Warn(" [WARNING] Received message with unknown routing key: %s", zap.String("routingKey", m.RoutingKey))
		return
	}

	if err != nil {
		m.Nack(false, true)
		return
	}

	m.Ack(false)
	logger.L().Info(" [INFO] Successfully processed event %s", zap.String("routingKey", m.RoutingKey))
}

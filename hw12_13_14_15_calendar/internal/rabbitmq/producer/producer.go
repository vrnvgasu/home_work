package producer

import (
	"fmt"

	"github.com/streadway/amqp"

	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/config"
)

type Producer struct {
	amqpURI      string
	exchange     string
	exchangeType string
	routingKey   string
}

func NewProducer() *Producer {
	return &Producer{
		amqpURI:      config.Cfg.Rabbitmq.URI,
		exchange:     config.Cfg.Rabbitmq.ExchangeName,
		exchangeType: config.Cfg.Rabbitmq.ExchangeType,
		routingKey:   config.Cfg.Rabbitmq.RoutingKey,
	}
}

func (p *Producer) Publish(messages []string) error {
	connection, err := amqp.Dial(p.amqpURI)
	if err != nil {
		return fmt.Errorf("producer Publish Dial failed: %w", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("producer Publish Channel failed: %w", err)
	}

	if err = channel.ExchangeDeclare(
		p.exchange,
		p.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("producer Publish ExchangeDeclare failed: %w", err)
	}

	for _, msg := range messages {
		if err = channel.Publish(
			p.exchange,
			p.routingKey,
			false,
			false,
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            []byte(msg),
				DeliveryMode:    amqp.Persistent,
				Priority:        0,
			},
		); err != nil {
			return fmt.Errorf("producer Publish Publish failed: %w", err)
		}
	}

	return nil
}

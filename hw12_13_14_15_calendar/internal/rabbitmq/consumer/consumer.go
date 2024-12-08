package consumer

import (
	"fmt"

	"github.com/streadway/amqp"

	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/config"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	queue   *amqp.Queue
}

func NewConsumer() (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     config.Cfg.Rabbitmq.ConsumerTag,
		queue:   nil,
	}

	var err error

	c.conn, err = amqp.Dial(config.Cfg.Rabbitmq.URI)
	if err != nil {
		return nil, fmt.Errorf("consumer NewConsumer Dial failed: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("consumer NewConsumer Channel failed: %w", err)
	}

	if err = c.channel.ExchangeDeclare(
		config.Cfg.Rabbitmq.ExchangeName,
		config.Cfg.Rabbitmq.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("consumer NewConsumer ExchangeDeclare failed: %w", err)
	}

	queue, err := c.channel.QueueDeclare(
		config.Cfg.Rabbitmq.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("consumer NewConsumer QueueDeclare failed: %w", err)
	}
	c.queue = &queue

	if err = c.channel.QueueBind(
		config.Cfg.Rabbitmq.Queue,
		config.Cfg.Rabbitmq.RoutingKey,
		config.Cfg.Rabbitmq.ExchangeName,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("consumer NewConsumer QueueBind failed: %w", err)
	}

	return c, nil
}

func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		c.queue.Name,
		c.tag,
		false,
		false,
		false,
		false,
		nil,
	)
}

func (c *Consumer) Shutdown() error {
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("consumer Shutdown Cancel failed: %w", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("consumer Shutdown Close failed: %w", err)
	}

	return nil
}

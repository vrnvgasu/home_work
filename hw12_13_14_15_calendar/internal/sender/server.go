package sender

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Consumer interface {
	Consume() (<-chan amqp.Delivery, error)
	Shutdown() error
}

type Sender struct {
	consumer Consumer
	logger   Logger
	done     chan error
}

func NewSender(consumer Consumer, logger Logger) *Sender {
	return &Sender{
		consumer: consumer,
		logger:   logger,
		done:     make(chan error),
	}
}

func (s *Sender) Run() error {
	deliveries, err := s.consumer.Consume()
	if err != nil {
		s.logger.Info("sender Run Consume: " + err.Error())

		return err
	}

	go s.Handle(deliveries)

	return nil
}

func (s *Sender) Handle(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		msg := fmt.Sprintf("got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		s.logger.Info(msg)
		d.Ack(true)
	}

	s.logger.Info("sender Handle: deliveries channel closed")

	s.done <- nil
}

func (s *Sender) Stop() {
	s.logger.Info("sender Stop")

	if err := s.consumer.Shutdown(); err != nil {
		s.logger.Error("sender Stop Shutdown: %s" + err.Error())
	}

	<-s.done
}

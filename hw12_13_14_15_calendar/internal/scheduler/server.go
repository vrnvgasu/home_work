package scheduler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/storage"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type App interface {
	EventList(ctx context.Context, params storage.Params) ([]storage.Event, error)
	EventListToSend(ctx context.Context) ([]storage.Event, error)
	SetEventsSent(ctx context.Context, eventIDs []uint64) error
	DeleteEventList(ctx context.Context, eventIDs []uint64) error
}

type Producer interface {
	Publish(messages []string) error
}

type Scheduler struct {
	logger         Logger
	producer       Producer
	app            App
	ticker         time.Duration
	eventsLifeTime time.Duration
}

func NewScheduler(producer Producer, logger Logger, app App) *Scheduler {
	return &Scheduler{
		logger:         logger,
		producer:       producer,
		app:            app,
		ticker:         time.Duration(config.Cfg.Scheduler.Ticker),
		eventsLifeTime: time.Duration(config.Cfg.Scheduler.EventsLifeTime),
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.ticker * time.Second) //nolint:durationcheck
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler Run Done: stop scheduler")

			return nil
		case <-ticker.C:
			err := s.sendEvents(ctx)
			if err != nil {
				s.logger.Info("scheduler Run sendEvents: " + err.Error())

				return err
			}
			if err = s.clearOldEvents(ctx); err != nil {
				s.logger.Error("scheduler Run clearOldEvents: " + err.Error())

				return err
			}
		}
	}
}

func (s *Scheduler) clearOldEvents(ctx context.Context) error {
	events, err := s.app.EventList(ctx, storage.Params{
		StartAtLEq: time.Now().Add(-1 * s.eventsLifeTime),
	})
	if err != nil {
		s.logger.Error("scheduler clearOldEvents EventList: " + err.Error())

		return err
	}

	if len(events) == 0 {
		return nil
	}

	ids := make([]uint64, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}

	if err = s.app.DeleteEventList(ctx, ids); err != nil {
		s.logger.Error("scheduler clearOldEvents DeleteEventList: " + err.Error())

		return err
	}

	return nil
}

func (s *Scheduler) sendEvents(ctx context.Context) error {
	events, err := s.app.EventListToSend(ctx)
	if err != nil {
		s.logger.Error("scheduler eventsToSend EventList: " + err.Error())

		return err
	}

	if len(events) == 0 {
		return nil
	}

	messages := make([]string, 0, len(events))
	for _, e := range events {
		message, err := json.Marshal(e)
		if err != nil {
			s.logger.Error("scheduler eventsToSend Marshal: " + err.Error())

			return err
		}

		messages = append(messages, string(message))
	}

	if err = s.producer.Publish(messages); err != nil {
		s.logger.Error("scheduler eventsToSend Publish: " + err.Error())

		return err
	}

	ids := make([]uint64, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}
	if err = s.app.SetEventsSent(ctx, ids); err != nil {
		s.logger.Error("scheduler eventsToSend SetEventsSent: " + err.Error())
	}

	return nil
}

package storage

import (
	"context"
	"time"
)

type Duration int

const (
	Day Duration = iota
	Week
	Month
)

type Params struct {
	Limit      int
	StartAtGEq time.Time `db:"start_at_g_eq"`
	StartAtLEq time.Time `db:"start_at_l_eq"`
	IsSent     *bool     `db:"is_sent"`
}

func (p *Params) SetStartAtLEqFromDuration(d Duration) *Params {
	if p.StartAtLEq.IsZero() {
		p.SetDefaultStartAtLEq()
	}

	switch d {
	case Day:
		p.StartAtLEq = p.StartAtLEq.AddDate(0, 0, 1)
	case Week:
		p.StartAtLEq = p.StartAtLEq.AddDate(0, 0, 7)
	case Month:
		p.StartAtLEq = p.StartAtLEq.AddDate(0, 1, 0)
	}

	return p
}

func (p *Params) SetDefaultStartAtLEq() *Params {
	if p.StartAtGEq.IsZero() {
		p.StartAtLEq = time.Now()
	} else {
		p.StartAtLEq = p.StartAtGEq
	}

	return p
}

type IStorage interface {
	Add(context.Context, Event) (uint64, error)
	Update(context.Context, Event) error
	Delete(ctx context.Context, eventIDs []uint64) error
	List(context.Context, Params) ([]Event, error)
	ListToSend(ctx context.Context) ([]Event, error)
	SetSent(ctx context.Context, eventIDs []uint64) error
	Connect(context.Context) error
	Close(context.Context) error
	Migrate() error
}

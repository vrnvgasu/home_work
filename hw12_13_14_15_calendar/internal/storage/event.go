package storage

import "time"

type Event struct {
	ID          uint64    `db:"id"`
	Title       string    `db:"title"`
	StartAt     time.Time `db:"start_at"`
	EndAt       time.Time `db:"end_at"`
	Description string    `db:"description"`
	OwnerID     uint64    `db:"owner_id"`
	SendBefore  int64     `db:"send_before"`
	IsSent      bool      `db:"is_sent"`
}

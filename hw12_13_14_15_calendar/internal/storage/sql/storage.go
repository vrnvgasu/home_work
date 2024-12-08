package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/stdlib" // postgresql provider
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/storage"
)

type listQueryBuilder struct {
	q            string
	isWhereExist bool
}

func newListQueryBuilder() *listQueryBuilder {
	return &listQueryBuilder{
		q: "select * from event",
	}
}

func (l *listQueryBuilder) where(condition string) {
	if !l.isWhereExist {
		l.q += " where "
	} else {
		l.q += " and "
	}

	l.q += condition + " "

	l.isWhereExist = true
}

func (l *listQueryBuilder) orderBy(condition string) {
	l.q += " order by " + condition + " "
}

func (l *listQueryBuilder) limit(limit int) {
	if limit > 0 {
		l.q += fmt.Sprintf(" limit %d ", limit)
	}
}

func (l *listQueryBuilder) build() string {
	return l.q
}

type Storage struct {
	db *sqlx.DB
}

func New() storage.IStorage {
	return &Storage{}
}

func (s *Storage) Add(ctx context.Context, e storage.Event) (uint64, error) {
	q := `insert into event (title, start_at, end_at, description, owner_id, send_before)
	values ($1, $2, $3, $4, $5, $6) returning id;`
	err := s.db.QueryRowxContext(ctx, q, e.Title, e.StartAt, e.EndAt, e.Description, e.OwnerID, e.SendBefore).
		Scan(&e.ID)
	if err != nil {
		return 0, fmt.Errorf("insert event error: %w", err)
	}

	return e.ID, nil
}

func (s *Storage) Update(ctx context.Context, e storage.Event) error {
	q := `update event
	set title = :title, start_at = :start_at,
    end_at = :end_at, description = :description, 
    owner_id = :owner_id, send_before = :send_before
	where id = :id;`
	_, err := s.db.NamedExecContext(ctx, q, e)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("id not found, %d: %w", e.ID, err)
	} else if err != nil {
		return fmt.Errorf("update event error: %w", err)
	}

	return nil
}

func (s *Storage) Delete(ctx context.Context, eventIDs []uint64) error {
	q := `delete from event where id = any($1)`
	_, err := s.db.ExecContext(ctx, q, pq.Array(eventIDs))
	if err != nil {
		return fmt.Errorf("delete event error: %w", err)
	}

	return nil
}

func (s *Storage) List(ctx context.Context, params storage.Params) ([]storage.Event, error) {
	qBuilder := newListQueryBuilder()

	if !params.StartAtGEq.IsZero() {
		qBuilder.where("start_at >= :start_at_g_eq")
	}

	if !params.StartAtLEq.IsZero() {
		qBuilder.where("start_at <= :start_at_l_eq")
	}

	if params.IsSent != nil {
		qBuilder.where("is_sent = :is_sent")
	}

	qBuilder.orderBy("start_at")
	qBuilder.limit(params.Limit)

	rows, err := s.db.NamedQueryContext(ctx, qBuilder.build(), params)
	if err != nil {
		return nil, fmt.Errorf("list event error: %w", err)
	}
	defer rows.Close()

	result := make([]storage.Event, 0)
	for rows.Next() {
		e := storage.Event{}
		if err = rows.StructScan(&e); err != nil {
			return nil, fmt.Errorf("list event scan error: %w", err)
		}

		result = append(result, e)
	}

	return result, nil
}

func (s *Storage) ListToSend(ctx context.Context) ([]storage.Event, error) {
	q := `
		select *
		from event e
		where e.start_at <= now() - e.send_before * interval '1 sec'
		and e.is_sent = false
		`

	rows, err := s.db.NamedQueryContext(ctx, q, struct{}{})
	if err != nil {
		return nil, fmt.Errorf("ListToSend event error: %w", err)
	}
	defer rows.Close()

	result := make([]storage.Event, 0)
	for rows.Next() {
		e := storage.Event{}
		if err = rows.StructScan(&e); err != nil {
			return nil, fmt.Errorf("ListToSend event scan error: %w", err)
		}

		result = append(result, e)
	}

	return result, nil
}

func (s *Storage) SetSent(ctx context.Context, eventIDs []uint64) error {
	q := `
		update event
		set is_sent = true
		where id = any($1)
		`
	_, err := s.db.ExecContext(ctx, q, pq.Array(eventIDs))
	if err != nil {
		return fmt.Errorf("SetSent event error: %w", err)
	}

	return nil
}

func (s *Storage) Connect(_ context.Context) error {
	db, err := sqlx.Open("pgx", config.Cfg.PSQL.DSN)
	if err != nil {
		return fmt.Errorf("open sql db fail: %w", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("ping sql db fail: %w", err)
	}

	s.db = db

	return nil
}

func (s *Storage) Close(_ context.Context) error {
	return s.db.Close()
}

func (s *Storage) Migrate() error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect failed: %w", err)
	}

	if err := goose.Up(s.db.DB, config.Cfg.PSQL.Migration); err != nil {
		return fmt.Errorf("up migration failed: %w", err)
	}

	return nil
}

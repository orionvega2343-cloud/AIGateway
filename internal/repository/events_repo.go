package repository

import (
	"AIGateway/internal/models"
	"context"

	"github.com/jmoiron/sqlx"
)

type Events interface {
	CreateEvent(ctx context.Context, m models.Event) (models.Event, error)
	GetEventById(ctx context.Context, id int) (models.Event, error)
}

type EventsRepo struct {
	db *sqlx.DB
}

func NewEvents(db *sqlx.DB) *EventsRepo {
	return &EventsRepo{db: db}
}

func (r *EventsRepo) CreateEvent(ctx context.Context, m models.Event) (models.Event, error) {
	//TODO: добавить created_at в RETURNING
	err := r.db.GetContext(ctx, &m, `INSERT INTO events(event_id, type, payload, status) VALUES($1, $2, $3, $4) RETURNING id`, m.EventId, m.Type, m.Payload, m.Status)
	if err != nil {
		return m, err
	}
	return m, nil
}

func (r *EventsRepo) GetEventById(ctx context.Context, id int) (models.Event, error) {
	var m models.Event
	err := r.db.GetContext(ctx, &m, `SELECT event_id, type, payload, status, created_at, processed_at FROM events WHERE id = $1`, id)
	if err != nil {
		return m, err
	}
	return m, nil
}

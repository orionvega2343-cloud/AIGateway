package repository

import (
	"AIGateway/internal/models"
	"context"

	"github.com/jmoiron/sqlx"
)

type Enrichments interface {
	Save(ctx context.Context, EventId int, content string, model string) (models.Enrichment, error)
	GetById(ctx context.Context, id int) (models.Enrichment, error)
}

type EnrichmentsRepo struct {
	db *sqlx.DB
}

func NewEnrichments(db *sqlx.DB) *EnrichmentsRepo {
	return &EnrichmentsRepo{db: db}
}

func (r *EnrichmentsRepo) Save(ctx context.Context, EventId int, content string, model string) (models.Enrichment, error) {
	var e models.Enrichment
	err := r.db.GetContext(ctx, &e, `INSERT INTO enrichments (event_id, response, model) VALUES($1, $2, $3) RETURNING id, created_at`, EventId, content, model)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (r *EnrichmentsRepo) GetById(ctx context.Context, id int) (models.Enrichment, error) {
	var e models.Enrichment
	err := r.db.GetContext(ctx, &e, `SELECT id, event_id, response,  model, created_at  FROM enrichments WHERE id = $1`, id)
	if err != nil {
		return e, err
	}
	return e, nil
}

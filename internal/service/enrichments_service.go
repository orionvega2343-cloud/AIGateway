package service

import (
	"AIGateway/internal/models"
	"AIGateway/internal/repository"
	"context"
)

type Enrichments interface {
	Save(ctx context.Context, EventId int, content string, model string) (models.Enrichment, error)
	GetById(ctx context.Context, id int) (models.Enrichment, error)
}

type EnrichmentsService struct {
	repo repository.Enrichments
}

func NewEnrichmentsService(repo repository.Enrichments) *EnrichmentsService {
	return &EnrichmentsService{repo: repo}
}

func (e *EnrichmentsService) Save(ctx context.Context, EventId int, content string, model string) (models.Enrichment, error) {
	return e.repo.Save(ctx, EventId, content, model)
}

func (e *EnrichmentsService) GetById(ctx context.Context, id int) (models.Enrichment, error) {
	return e.repo.GetById(ctx, id)
}

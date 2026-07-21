package service

import (
	"AIGateway/internal/models"
	"AIGateway/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Events interface {
	CreateEvent(ctx context.Context, externalId string, payload string) (models.Event, error)
	GetEventById(ctx context.Context, id int) (models.Event, error)
}

var ErrDuplicateEvent = errors.New("duplicate event")

type EventsService struct {
	repo        repository.Events
	redisClient *redis.Client
	expiration  time.Duration
}

func NewEventsService(repo repository.Events, redisClient *redis.Client, expiration time.Duration) *EventsService {
	return &EventsService{repo: repo, redisClient: redisClient, expiration: expiration}
}

func (e *EventsService) CreateEvent(ctx context.Context, externalId string, payload string) (models.Event, error) {
	//Если ключ уже создан,
	//ничего не делаем
	resp, err := e.redisClient.SetNX(ctx, "dedup:"+externalId, 1, e.expiration).Result()
	if err != nil {
		return models.Event{}, err
	}
	//Если false,
	//не сохраняем в БД
	if !resp {
		return models.Event{}, ErrDuplicateEvent
	}

	ev := models.Event{EventId: externalId, Payload: payload, Status: "pending"}
	event, err := e.repo.CreateEvent(ctx, ev)
	if err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (e *EventsService) GetEventById(ctx context.Context, id int) (models.Event, error) {
	return e.repo.GetEventById(ctx, id)
}

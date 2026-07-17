package models

import "time"

type Event struct {
	Id          int        `json:"id"`
	EventId     int        `json:"event_id"`
	Type        string     `json:"type"`
	Payload     string     `json:"payload"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at"`
}

package models

import "time"

type Event struct {
	Id          int        `json:"id" db:"id"`
	EventId     string     `json:"event_id" db:"event_id"`
	Type        string     `json:"type" db:"type"`
	Payload     string     `json:"payload" db:"payload"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ProcessedAt *time.Time `json:"processed_at" db:"processed_at"`
}

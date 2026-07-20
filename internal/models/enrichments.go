package models

import "time"

type Enrichment struct {
	Id        int       `json:"id" db:"id"`
	EventId   int       `json:"event_id" db:"event_id"`
	Response  string    `json:"response" db:"response"`
	Model     string    `json:"model" db:"model"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

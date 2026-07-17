package models

import "time"

type Enrichment struct {
	Id        int       `json:"id"`
	EventId   int       `json:"event_id"`
	Response  string    `json:"response"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

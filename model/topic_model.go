package model

import (
  "github.com/google/uuid"
  "time"
)

// Topic represents a subject that I write about.
type Topic struct {
  UUID      uuid.UUID `json:"uuid"`
  Name      string    `json:"name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

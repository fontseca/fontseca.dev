package model

import (
  "github.com/google/uuid"
  "time"
)

// Topic represents a subject that I write about.
type Topic struct {
  ID        uuid.UUID `json:"id"`
  Name      string    `json:"name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

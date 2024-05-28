package model

import (
  "time"
)

// Topic represents a subject that I write about.
type Topic struct {
  ID        string    `json:"id"`
  Name      string    `json:"name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

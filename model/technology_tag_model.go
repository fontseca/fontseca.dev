package model

import (
  "github.com/google/uuid"
  "time"
)

// TechnologyTag is a tag that helps further describe a project.
type TechnologyTag struct {
  UUID      uuid.UUID `json:"uuid"`
  Name      string    `json:"name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

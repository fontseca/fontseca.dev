package model

import (
  "github.com/google/uuid"
  "time"
)

// TechnologyTag is a tag that helps further describe a project.
type TechnologyTag struct {
  ID        uuid.UUID `json:"id"`
  Name      string    `json:"name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

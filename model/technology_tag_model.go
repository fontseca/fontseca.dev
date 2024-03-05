package model

import (
  "github.com/google/uuid"
)

// TechnologyTag is a tag that helps further describe a project.
type TechnologyTag struct {
  ID   uuid.UUID `json:"id"`
  Name string    `json:"name"`
}

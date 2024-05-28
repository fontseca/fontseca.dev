package model

import (
  "time"
)

// Tag is an entity attached to articles to provide more
// information about the topic or their content.
type Tag struct {
  ID        string    `json:"id"`
  Name      string    `json:"name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

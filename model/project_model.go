package model

import (
  "github.com/google/uuid"
  "time"
)

// Project represents a project that is being developed by me.
type Project struct {
  ID             uuid.UUID `json:"id"`
  Name           string    `json:"name"`
  Homepage       string    `json:"homepage"`
  Language       string    `json:"language"`
  Summary        string    `json:"summary"`
  Content        string    `json:"content"`
  EstimatedTime  int       `json:"estimated_time"`
  FirstImageURL  string    `json:"first_image_url"`
  SecondImageURL string    `json:"second_image_url"`
  GitHubURL      string    `json:"github_url"`
  CollectionURL  string    `json:"collection_url"`
  PlaygroundURL  string    `json:"playground_url"`
  Playable       bool      `json:"playable"`
  Archived       bool      `json:"archived"`
  Finished       bool      `json:"finished"`
  CreatedAt      time.Time `json:"created_at"`
  UpdatedAt      time.Time `json:"updated_at"`
}

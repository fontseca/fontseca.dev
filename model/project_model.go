package model

import (
  "github.com/google/uuid"
  "time"
)

// Project represents a project that is being developed by me.
type Project struct {
  UUID            uuid.UUID  `json:"uuid"`
  Name            string     `json:"name"`
  Slug            string     `json:"slug"`
  Homepage        string     `json:"homepage"`
  Company         *string    `json:"company"`
  CompanyHomepage *string    `json:"company_homepage"`
  Starts          *time.Time `json:"starts"`
  Ends            *time.Time `json:"ends"`
  Language        *string    `json:"language"`
  Summary         string     `json:"summary"`
  ReadTime        int        `json:"read_time"`
  Content         string     `json:"content"`
  FirstImageURL   string     `json:"first_image_url"`
  SecondImageURL  string     `json:"second_image_url"`
  GitHubURL       string     `json:"github_url"`
  CollectionURL   string     `json:"collection_url"`
  PlaygroundURL   string     `json:"playground_url"`
  Playable        bool       `json:"playable"`
  Archived        bool       `json:"archived"`
  Finished        bool       `json:"finished"`
  TechnologyTags  []string   `json:"technology_tags"`
  CreatedAt       time.Time  `json:"created_at"`
  UpdatedAt       time.Time  `json:"updated_at"`
}

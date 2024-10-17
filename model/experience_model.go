package model

import (
  "github.com/google/uuid"
  "time"
)

// Experience represents a work experience entry.
type Experience struct {
  UUID            uuid.UUID `json:"uuid"`
  Starts          int       `json:"starts"`
  Ends            *int      `json:"ends"`
  JobTitle        string    `json:"job_title"`
  Company         string    `json:"company"`
  CompanyHomepage *string   `json:"company_homepage"`
  Country         string    `json:"country"`
  Summary         string    `json:"summary"`
  Active          bool      `json:"active"`
  Hidden          bool      `json:"hidden"`
  CreatedAt       time.Time `json:"created_at"`
  UpdatedAt       time.Time `json:"updated_at"`
}

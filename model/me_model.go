package model

import (
  "time"
)

// Me contains information about me. Such as contact links and any
// other necessary metadata.
type Me struct {
  Username     string    `json:"username"`
  FirstName    string    `json:"first_name"`
  LastName     string    `json:"last_name"`
  Summary      string    `json:"summary"`
  JobTitle     string    `json:"job_title"`
  Email        string    `json:"email"`
  PhotoURL     string    `json:"photo_url"`
  ResumeURL    string    `json:"resume_url"`
  CodingSince  int       `json:"coding_since"`
  Company      string    `json:"company"`
  Location     string    `json:"location"`
  Hireable     bool      `json:"hireable"`
  GitHubURL    string    `json:"github_url"`
  LinkedInURL  string    `json:"linkedin_url"`
  YouTubeURL   string    `json:"youtube_url"`
  TwitterURL   string    `json:"twitter_url"`
  InstagramURL string    `json:"instagram_url"`
  CreatedAt    time.Time `json:"created_at"`
  UpdatedAt    time.Time `json:"updated_at"`
}

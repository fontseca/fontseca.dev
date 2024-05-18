package model

import (
  "github.com/google/uuid"
  "time"
)

// Article represents a piece of writing about a
// particular subject in my website's archive.
type Article struct {
  UUID        uuid.UUID  `json:"uuid"`
  Title       string     `json:"title"`
  Author      string     `json:"author"`
  Slug        string     `json:"slug"`
  ReadTime    int        `json:"read_time"`
  Content     string     `json:"content"`
  Topics      []*Topic   `json:"topics"`
  IsDraft     bool       `json:"is_draft"`
  IsPinned    bool       `json:"is_pinned"`
  PublishedAt *time.Time `json:"published_at"`
  ModifiedAt  *time.Time `json:"modified_at"`
  DraftedAt   time.Time  `json:"drafted_at"`
  UpdatedAt   time.Time  `json:"updated_at"`
}

// ArticlePatch is a patch for a published article.
type ArticlePatch struct {
  ArticleUUID uuid.UUID `json:"article_uuid"`
  Title       *string   `json:"title"`
  Slug        *string   `json:"slug"`
  ReadTime    *int
  Content     *string `json:"content"`
}

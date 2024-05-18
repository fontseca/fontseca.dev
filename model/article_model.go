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
  IsDraft     time.Time  `json:"is_draft"`
  IsPinned    time.Time  `json:"is_pinned"`
  DraftedAt   time.Time  `json:"drafted_at"`
  PinnedAt    *time.Time `json:"pinned_at"`
  ArchivedAt  *time.Time `json:"archived_at"`
  PublishedAt *time.Time `json:"published_at"`
  ModifiedAt  *time.Time `json:"modified_at"`
  CreatedAt   time.Time  `json:"created_at"`
  UpdatedAt   time.Time  `json:"updated_at"`
}

// ArticlePatch is a patch for a published article.
type ArticlePatch struct {
  ArticleUUID uuid.UUID `json:"article_uuid"`
  Title       *string   `json:"title"`
  Slug        *string   `json:"slug"`
  Content     *string   `json:"content"`
}

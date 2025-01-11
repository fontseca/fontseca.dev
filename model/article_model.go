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
  Slug        string     `json:"slug"`
  Author      string     `json:"author"`
  Views       int64      `json:"views"`
  ReadTime    int        `json:"read_time"`
  IsDraft     bool       `json:"is_draft"`
  IsPinned    bool       `json:"is_pinned"`
  PublishedAt *time.Time `json:"published_at"`
  ModifiedAt  *time.Time `json:"modified_at"`
  DraftedAt   time.Time  `json:"drafted_at"`
  UpdatedAt   time.Time  `json:"updated_at"`
  Topic       *Topic     `json:"topic"`
  Tags        []*Tag     `json:"tags"`
  Summary     string     `json:"summary"`
  CoverURL    string     `json:"cover_url"`
  CoverCap    *string    `json:"cover_caption"`
  Content     string     `json:"content"`
}

// ArticlePatch is a patch for a published article.
type ArticlePatch struct {
  ArticleUUID uuid.UUID `json:"article_uuid"`
  Title       *string   `json:"title"`
  Slug        *string   `json:"slug"`
  ReadTime    *int      `json:"-"`
  TopicID     *string   `json:"topic_id"`
  Content     *string   `json:"content"`
}

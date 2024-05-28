package transfer

import (
  "github.com/google/uuid"
  "time"
)

// ArticleCreation represents the data required to create a new article entry.
type ArticleCreation struct {
  Title    string `json:"title" binding:"required,max=256"`
  Slug     string
  ReadTime int
  Content  string `json:"content"`
}

// ArticleUpdate represents the data required to update an existing article entry.
type ArticleUpdate struct {
  Title    string `json:"title"`
  Slug     string
  ReadTime int
  Content  string `json:"content"`
}

// Article is a shallow article entry for transferring metadata.
type Article struct {
  UUID        uuid.UUID `json:"uuid"`
  Title       string    `json:"title"`
  URL         string    `json:"url"` // in the form: 'https://fontseca.dev/archive/:topic/:year/:month/:slug'
  IsPinned    bool      `json:"is_pinned"`
  PublishedAt time.Time `json:"published_at"`
}

// Publication represents the publication date of an article, consisting of a month and a year.
type Publication struct {
  Month time.Month
  Year  int
}

// ArticleFilter represents the parameters used to query articles.
type ArticleFilter struct {
  Search      string
  Publication *Publication
  IsHidden    bool
  IsDraft     bool
  Page        int
  RPP         int // records per page
}

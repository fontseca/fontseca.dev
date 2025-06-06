package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "log/slog"
  "strings"
)

type archiveRepositoryAPIForDrafts interface {
  Draft(ctx context.Context, creation *transfer.ArticleCreation) (draft string, err error)
  Publish(ctx context.Context, draftID string) error
  List(ctx context.Context, filter *transfer.ArticleFilter, hidden, draftsOnly bool) (drafts []*transfer.Article, err error)
  GetByLink(ctx context.Context, link string) (article *model.Article, err error)
  GetByID(ctx context.Context, draftID string, isDraft bool) (draft *model.Article, err error)
  AddTag(ctx context.Context, draftID, tagID string, isDraft ...bool) error
  RemoveTag(ctx context.Context, draftID, tagID string, isDraft ...bool) error
  Share(ctx context.Context, draftID string) (link string, err error)
  Discard(ctx context.Context, draftID string) error
  Revise(ctx context.Context, draftID string, revision *transfer.ArticleRevision) error
}

// DraftsService is a high level provider for article drafts.
type DraftsService struct {
  r archiveRepositoryAPIForDrafts
}

func NewDraftsService(r archiveRepositoryAPIForDrafts) *DraftsService {
  return &DraftsService{r}
}

// Draft starts the creation process of an article. It returns the
// UUID of the draft that was created.
//
// To draft an article, only its title is required, other fields
// are completely optional and can be added in an eventual revision.
func (s *DraftsService) Draft(ctx context.Context, creation *transfer.ArticleCreation) (insertedUUID uuid.UUID, err error) {
  if nil == creation {
    err = errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return uuid.Nil, err
  }

  creation.Title = strings.TrimSpace(creation.Title)
  creation.Summary = strings.TrimSpace(creation.Summary)
  creation.Content = strings.TrimSpace(creation.Content)
  creation.CoverURL = strings.TrimSpace(creation.CoverURL)
  creation.CoverCap = strings.TrimSpace(creation.CoverCap)

  sanitizeTextWordIntersections(&creation.Title)

  if "" != creation.CoverURL {
    err = sanitizeURL(&creation.CoverURL)
    if nil != err {
      return uuid.Nil, err
    }
  }

  switch {
  case 256 < len(creation.Title):
    return uuid.Nil, problem.NewValidation([3]string{"title", "max", "256"})
  case 0 != len(creation.Content) && 3145728 < len(creation.Content):
    return uuid.Nil, problem.NewValidation([3]string{"content", "max", "3145728"})
  case 0 != len(creation.Summary) && 512 < len(creation.Summary):
    return uuid.Nil, problem.NewValidation([3]string{"summary", "max", "512"})
  case 0 != len(creation.CoverCap) && 256 < len(creation.CoverCap):
    return uuid.Nil, problem.NewValidation([3]string{"cover_caption", "max", "256"})
  }

  creation.Slug = generateSlug(creation.Title)

  builder := strings.Builder{}

  builder.WriteString(creation.Title)

  if "" != creation.Summary {
    builder.WriteRune('\n')
    builder.WriteString(creation.Summary)
  }

  if "" != creation.CoverCap {
    builder.WriteRune('\n')
    builder.WriteString(creation.CoverCap)
  }

  if "" != creation.Content {
    builder.WriteRune('\n')
    builder.WriteString(creation.Content)
  }

  reader := strings.NewReader(builder.String())
  creation.ReadTime = computePostReadingTimeInMinutes(reader)

  id, err := s.r.Draft(ctx, creation)
  if err != nil {
    return uuid.Nil, err
  }

  return uuid.Parse(id)
}

// Publish makes a draft publicly available.
//
// Invoking Publish on an already published article has no effect.
func (s *DraftsService) Publish(ctx context.Context, draftUUID string) error {
  if err := validateUUID(&draftUUID); nil != err {
    return err
  }

  return s.r.Publish(ctx, draftUUID)
}

// List retrieves all the ongoing articles drafts.
//
// If [filter.Search] is a non-empty string, then Get behaves like a search
// function over draft articles, so it attempts to find and
// amass every article whose title contains any of the keywords
// (if more than one) in filter.Search.
func (s *DraftsService) List(ctx context.Context, filter *transfer.ArticleFilter) (drafts []*transfer.Article, err error) {
  return s.r.List(ctx, filter, false, true)
}

// GetByLink retrieves a draft by its shareable link.
func (s *DraftsService) GetByLink(ctx context.Context, link string) (article *model.Article, err error) {
  return s.r.GetByLink(ctx, link)
}

// Get retrieves one article draft by its UUID.
func (s *DraftsService) Get(ctx context.Context, draftUUID string) (draft *model.Article, err error) {
  if err = validateUUID(&draftUUID); nil != err {
    return nil, err
  }

  return s.r.GetByID(ctx, draftUUID, true)
}

// AddTag adds a tag to the article draft. If the tag already
// exists, it returns an error informing about a conflicting state.
func (s *DraftsService) AddTag(ctx context.Context, draftUUID, tagID string) error {
  if err := validateUUID(&draftUUID); nil != err {
    return err
  }

  return s.r.AddTag(ctx, draftUUID, tagID, true)
}

// RemoveTag removes a tag from the article draft. If the article
// has  no tag identified by its UUID, it returns an error indication
// a not found state.
func (s *DraftsService) RemoveTag(ctx context.Context, draftUUID, tagID string) error {
  if err := validateUUID(&draftUUID); nil != err {
    return err
  }

  return s.r.RemoveTag(ctx, draftUUID, tagID, true)
}

// Share creates a shareable link for an article draft. Only users
// with that link can see the progress and provide feedback.
//
// A shareable link does not make an article public. This link will
// eventually expire after a certain amount of time.
func (s *DraftsService) Share(ctx context.Context, draftUUID string) (link string, err error) {
  if err = validateUUID(&draftUUID); nil != err {
    return "about:blank", err
  }

  link, err = s.r.Share(ctx, draftUUID)

  if nil != err {
    return "about:blank", err
  }

  return link, nil
}

// Discard completely drops an article draft.
func (s *DraftsService) Discard(ctx context.Context, draftUUID string) error {
  if err := validateUUID(&draftUUID); nil != err {
    return err
  }

  return s.r.Discard(ctx, draftUUID)
}

// Revise adds a correction or inclusion to an article draft in order
// to correct or improve it.
func (s *DraftsService) Revise(ctx context.Context, draftUUID string, revision *transfer.ArticleRevision) error {
  if nil == revision {
    err := errors.New("nil value for parameter: revision")
    slog.Error(err.Error())
    return err
  }

  if err := validateUUID(&draftUUID); nil != err {
    return err
  }

  revision.Title = strings.TrimSpace(revision.Title)
  revision.Content = strings.TrimSpace(revision.Content)

  if "" != revision.Title {
    sanitizeTextWordIntersections(&revision.Title)
  }

  if "" != revision.CoverURL {
    err := sanitizeURL(&revision.CoverURL)
    if nil != err {
      return err
    }
  }

  switch {
  case 0 != len(revision.Title) && 256 < len(revision.Title):
    return problem.NewValidation([3]string{"title", "max", "256"})
  case 0 != len(revision.Content) && 3145728 < len(revision.Content):
    return problem.NewValidation([3]string{"content", "max", "3145728"})
  case 0 != len(revision.Summary) && 512 < len(revision.Summary):
    return problem.NewValidation([3]string{"summary", "max", "512"})
  case 0 != len(revision.Summary) && 100 > len(revision.Summary):
    return problem.NewValidation([3]string{"summary", "min", "120"})
  case 0 != len(revision.CoverCap) && 256 < len(revision.CoverCap):
    return problem.NewValidation([3]string{"cover_caption", "max", "256"})
  }

  if "" != revision.Title {
    revision.Slug = generateSlug(revision.Title)
  }

  if "" != revision.Title || "" != revision.Content || "" != revision.Summary || "" != revision.CoverCap {
    current, err := s.Get(ctx, draftUUID)
    if nil != err {
      return err
    }

    var (
      title        = revision.Title
      content      = revision.Content
      summary      = revision.Summary
      coverCaption = revision.CoverCap
    )

    if "" == title && "" != current.Title {
      title = current.Title
    }

    if "" == content && "" != current.Content {
      content = current.Content
    }

    if "" == summary && "" != current.Summary {
      summary = current.Summary
    }

    if "" == coverCaption && nil != current.CoverCap {
      coverCaption = *current.CoverCap
    }

    builder := strings.Builder{}

    /* None of these variables should be an empty string (except cover caption), but for correctness,
       I add the 'if' wrappers.  */

    if "" != title {
      builder.WriteString(title)
    }

    if "" != content {
      builder.WriteRune('\n')
      builder.WriteString(content)
    }

    if "" != summary {
      builder.WriteRune('\n')
      builder.WriteString(summary)
    }

    if "" != coverCaption {
      builder.WriteRune('\n')
      builder.WriteString(summary)
    }

    r := strings.NewReader(builder.String())
    revision.ReadTime = computePostReadingTimeInMinutes(r)
  }

  return s.r.Revise(ctx, draftUUID, revision)
}

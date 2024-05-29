package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/repository"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "log/slog"
  "strings"
)

// DraftsService is a high level provider for article drafts.
type DraftsService interface {
  // Draft starts the creation process of an article. It returns the
  // UUID of the draft that was created.
  //
  // To draft an article, only its title is required, other fields
  // are completely optional and can be added in an eventual revision.
  Draft(ctx context.Context, creation *transfer.ArticleCreation) (insertedUUID uuid.UUID, err error)

  // Publish makes a draft publicly available.
  //
  // Invoking Publish on an already published article has no effect.
  Publish(ctx context.Context, draftUUID string) error

  // Get retrieves all the ongoing articles drafts.
  //
  // If filter.Search is a non-empty string, then Get behaves like a search
  // function over draft articles, so it attempts to find and
  // amass every article whose title contains any of the keywords
  // (if more than one) in filter.Search.
  Get(ctx context.Context, filter *transfer.ArticleFilter) (drafts []*transfer.Article, err error)

  // GetByID retrieves one article draft by its UUID.
  GetByID(ctx context.Context, draftUUID string) (draft *model.Article, err error)

  // AddTag adds a tag to the article draft. If the tag already
  // exists, it returns an error informing about a conflicting state.
  AddTag(ctx context.Context, draftUUID, tagID string) error

  // RemoveTag removes a tag from the article draft. If the article
  // has  no tag identified by its UUID, it returns an error indication
  // a not found state.
  RemoveTag(ctx context.Context, draftUUID, tagID string) error

  // Share creates a shareable link for an article draft. Only users
  // with that link can see the progress and provide feedback.
  //
  // A shareable link does not make an article public. This link will
  // eventually expire after a certain amount of time.
  Share(ctx context.Context, draftUUID string) (link string, err error)

  // Discard completely drops an article draft.
  Discard(ctx context.Context, draftUUID string) error

  // Revise adds a correction or inclusion to an article draft in order
  // to correct or improve it.
  Revise(ctx context.Context, draftUUID string, revision *transfer.ArticleUpdate) error
}

type draftsService struct {
  r repository.ArchiveRepository
}

func NewDraftsService(r repository.ArchiveRepository) DraftsService {
  return &draftsService{r}
}

func (s *draftsService) Draft(ctx context.Context, creation *transfer.ArticleCreation) (insertedUUID uuid.UUID, err error) {
  if nil == creation {
    err = errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return uuid.Nil, err
  }

  creation.Title = strings.TrimSpace(creation.Title)

  sanitizeTextWordIntersections(&creation.Title)

  switch {
  case 256 < len(creation.Title):
    return uuid.Nil, problem.NewValidation([3]string{"title", "max", "256"})
  case 0 != len(creation.Content) && 3145728 < len(creation.Content):
    return uuid.Nil, problem.NewValidation([3]string{"content", "max", "3145728"})
  }

  creation.Slug = generateSlug(creation.Title)

  builder := strings.Builder{}

  builder.WriteString(creation.Title)

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

func (s *draftsService) Publish(ctx context.Context, draftUUID string) error {
  if err := validateUUID(&draftUUID); nil != err {
    return err
  }

  return s.r.Publish(ctx, draftUUID)
}

func (s *draftsService) Get(ctx context.Context, filter *transfer.ArticleFilter) (drafts []*transfer.Article, err error) {
  return s.r.Get(ctx, filter, false, true)
}

func (s *draftsService) GetByID(ctx context.Context, draftUUID string) (draft *model.Article, err error) {
  if err = validateUUID(&draftUUID); nil != err {
    return nil, err
  }

  return s.r.GetByID(ctx, draftUUID, true)
}

func (s *draftsService) AddTag(ctx context.Context, draftUUID, tagID string) error {
  if err := validateUUID(&draftUUID); nil != err {
    return err
  }

  return s.r.AddTag(ctx, draftUUID, tagID, true)
}

func (s *draftsService) RemoveTag(ctx context.Context, draftUUID, tagID string) error {
  if err := validateUUID(&draftUUID); nil != err {
    return err
  }

  return s.r.RemoveTag(ctx, draftUUID, tagID, true)
}

func (s *draftsService) Share(ctx context.Context, draftUUID string) (link string, err error) {
  if err = validateUUID(&draftUUID); nil != err {
    return "about:blank", err
  }

  link, err = s.r.Share(ctx, draftUUID)

  if nil != err {
    return "about:blank", err
  }

  return link, nil
}

func (s *draftsService) Discard(ctx context.Context, draftUUID string) error {
  if err := validateUUID(&draftUUID); nil != err {
    return err
  }

  return s.r.Discard(ctx, draftUUID)
}

func (s *draftsService) Revise(ctx context.Context, draftUUID string, revision *transfer.ArticleUpdate) error {
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

  switch {
  case 0 != len(revision.Title) && 256 < len(revision.Title):
    return problem.NewValidation([3]string{"title", "max", "256"})
  case 0 != len(revision.Content) && 3145728 < len(revision.Content):
    return problem.NewValidation([3]string{"content", "max", "3145728"})
  }

  if "" != revision.Title || "" != revision.Content {
    builder := strings.Builder{}

    if "" != revision.Title {
      builder.WriteString(revision.Title)
      revision.Slug = generateSlug(revision.Title)
    }

    if "" != revision.Content {
      if 0 < builder.Len() {
        builder.WriteRune('\n')
      }

      builder.WriteString(revision.Content)
    }

    r := strings.NewReader(builder.String())
    revision.ReadTime = computePostReadingTimeInMinutes(r)
  }

  return s.r.Revise(ctx, draftUUID, revision)
}

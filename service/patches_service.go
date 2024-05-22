package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/repository"
  "fontseca.dev/transfer"
  "log/slog"
  "strings"
)

// PatchesService is a high level provider for article patches.
type PatchesService interface {
  // Get retrieves all the ongoing article patches.
  Get(ctx context.Context) (patches []*model.ArticlePatch, err error)

  // Revise adds a correction or inclusion to an article patch in order
  // to correct or improve it.
  Revise(ctx context.Context, id string, revision *transfer.ArticleUpdate) error

  // Share creates a shareable link for an article patch. Only users
  // with that link can see the progress and provide feedback.
  //
  // A shareable link does not make an article public. This link will
  // eventually expire after a certain amount of time.
  Share(ctx context.Context, id string) (link string, err error)

  // Discard completely drops an article patch but keeps the original
  // article.
  Discard(ctx context.Context, id string) error

  // Release merges a patch into the original article and published the
  // update immediately after merging.
  Release(ctx context.Context, id string) error
}

type patchesService struct {
  r repository.ArchiveRepository
}

func NewPatchesService(r repository.ArchiveRepository) PatchesService {
  return &patchesService{r}
}

func (s *patchesService) Get(ctx context.Context) (patches []*model.ArticlePatch, err error) {
  return s.r.GetPatches(ctx)
}

func (s *patchesService) Revise(ctx context.Context, id string, revision *transfer.ArticleUpdate) error {
  if nil == revision {
    err := errors.New("nil value for parameter: revision")
    slog.Error(err.Error())
    return err
  }

  if err := validateUUID(&id); nil != err {
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

  return s.r.Revise(ctx, id, revision)
}

func (s *patchesService) Share(ctx context.Context, id string) (link string, err error) {
  if err = validateUUID(&id); nil != err {
    return "about:blank", err
  }

  link, err = s.r.Share(ctx, id)

  if nil != err {
    return "about:blank", err
  }

  return link, nil
}

func (s *patchesService) Discard(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.Discard(ctx, id)
}

func (s *patchesService) Release(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.Release(ctx, id)
}

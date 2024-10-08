package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "log/slog"
  "strings"
)

type archiveRepositoryAPIForPatches interface {
  ListPatches(ctx context.Context) (patches []*model.ArticlePatch, err error)
  Revise(ctx context.Context, patchID string, revision *transfer.ArticleRevision) error
  Share(ctx context.Context, patchID string) (link string, err error)
  Discard(ctx context.Context, patchID string) error
  Release(ctx context.Context, patchID string) error
}

// PatchesService is a high level provider for article patches.
type PatchesService struct {
  r archiveRepositoryAPIForPatches
}

func NewPatchesService(r archiveRepositoryAPIForPatches) *PatchesService {
  return &PatchesService{r}
}

// List retrieves all the ongoing article patches.
func (s *PatchesService) List(ctx context.Context) (patches []*model.ArticlePatch, err error) {
  return s.r.ListPatches(ctx)
}

// Revise adds a correction or inclusion to an article patch in order
// to correct or improve it.
func (s *PatchesService) Revise(ctx context.Context, id string, revision *transfer.ArticleRevision) error {
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

// Share creates a shareable link for an article patch. Only users
// with that link can see the progress and provide feedback.
//
// A shareable link does not make an article public. This link will
// eventually expire after a certain amount of time.
func (s *PatchesService) Share(ctx context.Context, id string) (link string, err error) {
  if err = validateUUID(&id); nil != err {
    return "about:blank", err
  }

  link, err = s.r.Share(ctx, id)

  if nil != err {
    return "about:blank", err
  }

  return link, nil
}

// Discard completely drops an article patch but keeps the original
// article.
func (s *PatchesService) Discard(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.Discard(ctx, id)
}

// Release merges a patch into the original article and published the
// update immediately after merging.
func (s *PatchesService) Release(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.Release(ctx, id)
}

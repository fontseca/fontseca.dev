package service

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
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

package service

import (
  "context"
  "fontseca.dev/model"
)

// ArticlesService is a high level provider for articles.
type ArticlesService interface {
  // Get retrieves all the published articles.
  //
  // If needle is a non-empty string, then Get behaves like a search
  // function over articles, so it attempts to find and amass every
  // article whose title contains any of the keywords (if more than one)
  // in needle.
  Get(ctx context.Context, needle string) (articles []*model.Article, err error)

  // GetByID retrieves one article by its UUID.
  GetByID(ctx context.Context, articleUUID string) (article *model.Article, err error)

  // Hide hides an article.
  Hide(ctx context.Context, id string) error

  // Show shows a hidden article.
  Show(ctx context.Context, id string) error

  // Amend starts the process to update an article. To amend the article,
  // a public copy of it is kept available to everyone while a patch
  // is created to store any revision made to the article.
  //
  // If the article is already being amended, any call to this method has
  // no effect.
  Amend(ctx context.Context, id string) error

  // Remove completely removes an article and any patch it currently has.
  Remove(ctx context.Context, id string) error

  // Pin pins an article.
  Pin(ctx context.Context, id string) error

  // Unpin unpins a pinned article.
  Unpin(ctx context.Context, id string) error

  // AddTopic adds a topic to the article. If the topic already
  // exists, it returns an error informing about a conflicting state.
  AddTopic(ctx context.Context, articleUUID, topicUUID string) error

  // RemoveTopic removes a topic from article draft. If the article
  // has no topic identified by its UUID, it returns an error indication
  // a not found state.
  RemoveTopic(ctx context.Context, articleUUID, topicUUID string) error
}

package service

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
)

// TopicsService is a high level provider for articles.
type TopicsService interface {
  // Add adds a new topic.
  Add(ctx context.Context, creation *transfer.TopicCreation) (insertedUUID string, err error)

  // Get retrieves all the topics.
  Get(ctx context.Context) (topics []*model.Topic, err error)

  // Update updates an existing topic.
  Update(ctx context.Context, id string, update *transfer.TopicUpdate) error

  // Remove removes a topic and detaches it from any article that uses it.
  Remove(ctx context.Context, id string) error
}

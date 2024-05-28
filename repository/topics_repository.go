package repository

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
)

// TopicsRepository is a low level API that provides methods for interacting
// with topics in the database.
type TopicsRepository interface {
  // Add adds a new topic.
  Add(ctx context.Context, creation *transfer.TopicCreation) error

  // Get retrieves all the topics.
  Get(ctx context.Context) (topics []*model.Topic, err error)

  // Update updates an existing topic.
  Update(ctx context.Context, id string, update *transfer.TopicUpdate) error

  // Remove removes a topic and detaches it from any article that uses it.
  Remove(ctx context.Context, id string) error
}

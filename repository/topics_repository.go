package repository

import (
  "context"
  "database/sql"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
)

// TopicsRepository is a low level API that provides methods for interacting
// with topics in the database.
type TopicsRepository interface {
  // Add adds a new topic.
  Add(ctx context.Context, creation *transfer.TopicCreation) (insertedUUID string, err error)

  // Get retrieves all the topics.
  Get(ctx context.Context) (topics []*model.Topic, err error)

  // Update updates an existing topic.
  Update(ctx context.Context, id string, update *transfer.TopicUpdate) error

  // Remove removes a topic and detaches it from any article that uses it.
  Remove(ctx context.Context, id string) error
}

type topicsRepository struct {
  db *sql.DB
}

func NewTopicsRepository(db *sql.DB) TopicsRepository {
  return &topicsRepository{db}
}

func (r *topicsRepository) Add(ctx context.Context, creation *transfer.TopicCreation) (insertedUUID string, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *topicsRepository) Get(ctx context.Context) (topics []*model.Topic, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *topicsRepository) Update(ctx context.Context, id string, update *transfer.TopicUpdate) error {
  // TODO implement me
  panic("implement me")
}

func (r *topicsRepository) Remove(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

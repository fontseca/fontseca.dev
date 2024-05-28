package repository

import (
  "context"
  "database/sql"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "log/slog"
  "net/http"
  "time"
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

type topicsRepository struct {
  db *sql.DB
}

func NewTopicsRepository(db *sql.DB) TopicsRepository {
  return &topicsRepository{db}
}

func (r *topicsRepository) Add(ctx context.Context, creation *transfer.TopicCreation) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    return err
  }

  defer tx.Rollback()

  addTopicQuery := `
  INSERT INTO "topic" ("id", "name")
               VALUES (@id, @name);`

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, addTopicQuery,
    sql.Named("id", creation.ID),
    sql.Named("name", creation.Name),
  )

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    p := problem.Problem{}
    p.Title("Topic not created.")
    p.Detail("Could not create topic for an unknown reason.")
    p.Status(http.StatusInternalServerError)

    return &p
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *topicsRepository) Get(ctx context.Context) (topics []*model.Topic, err error) {
  getTopicsQuery := `
  SELECT "id",
         "name",
         "created_at",
         "updated_at"
    FROM "topic"
ORDER BY lower("name");`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx, getTopicsQuery)
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  defer result.Close()

  topics = make([]*model.Topic, 0)

  for result.Next() {
    var topic model.Topic

    err = result.Scan(
      &topic.ID,
      &topic.Name,
      &topic.CreatedAt,
      &topic.UpdatedAt,
    )

    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }

    topics = append(topics, &topic)
  }

  return topics, nil
}

func (r *topicsRepository) Update(ctx context.Context, id string, update *transfer.TopicUpdate) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  updateArticleTopicQuery := `
  UPDATE "article"
     SET "topic" = @new_topic_id
   WHERE "topic" = @topic_id;`

  ctx1, cancel := context.WithTimeout(ctx, 7*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx1, updateArticleTopicQuery,
    sql.Named("topic_id", id),
    sql.Named("new_topic_id", update.ID),
  )

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  updateTopicQuery := `
  UPDATE "topic"
     SET "id" = @new_topic_id,
         "name" = @name,
         "updated_at" = current_timestamp
   WHERE "id" = @id;`

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err = tx.ExecContext(ctx, updateTopicQuery,
    sql.Named("id", id),
    sql.Named("new_topic_id", update.ID),
    sql.Named("name", update.Name),
  )

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "topic")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *topicsRepository) Remove(ctx context.Context, id string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  removeTopicQuery := `
  DELETE FROM "topic"
        WHERE "id" = $1;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx1, removeTopicQuery, id)

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "topic")
  }

  removeFromAttachedArticlesQuery := `
  UPDATE "article"
     SET "topic" = NULL
   WHERE "topic" = $1;`

  ctx1, cancel = context.WithTimeout(ctx, 7*time.Second)
  defer cancel()

  result, err = tx.ExecContext(ctx1, removeFromAttachedArticlesQuery, id)

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if _, err = result.RowsAffected(); nil != err {
    slog.Error(err.Error())
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

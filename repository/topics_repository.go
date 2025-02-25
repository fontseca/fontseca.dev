package repository

import (
  "context"
  "database/sql"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "log/slog"
  "net/http"
  "strings"
  "time"
)

// TopicsRepository is a low level API that provides methods for interacting
// with topics in the database.
type TopicsRepository struct {
  db *sql.DB
}

func NewTopicsRepository(db *sql.DB) *TopicsRepository {
  return &TopicsRepository{db}
}

// Create adds a new topic.
func (r *TopicsRepository) Create(ctx context.Context, creation *transfer.TopicCreation) error {
  slog.Info("adding new article topic",
    slog.String("id", creation.ID),
    slog.String("name", creation.Name))

  exists := false
  err := r.db.QueryRowContext(ctx, `SELECT count (1) FROM "archive"."topic" WHERE "id" = $1;`, creation.ID).Scan(&exists)

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if exists {
    p := &problem.Problem{}
    p.Status(http.StatusConflict)
    p.Type(problem.TypeDuplicateKey)
    p.Title("Could not create topic.")
    p.Detail("This topic is already registered.")
    p.With("topic_id", creation.ID)

    return p
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    return err
  }

  defer tx.Rollback()

  addTopicQuery := `
  INSERT INTO "archive"."topic" ("id", "name")
               VALUES ($1, $2);`

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, addTopicQuery, creation.ID, creation.Name)

  if nil != err {
    slog.Error(getErrMsg(err))
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
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// List retrieves all the topics.
func (r *TopicsRepository) List(ctx context.Context) (topics []*model.Topic, err error) {
  getTopicsQuery := `
  SELECT t."id",
         t."name",
         t."created_at",
         t."updated_at"
    FROM "archive"."topic" t
   WHERE 1 <= (SELECT count(a.*)
                 FROM "archive"."article" a
                WHERE NOT a."hidden"
                      AND a."published_at" IS NOT NULL
                      AND a."topic" = t."id")
ORDER BY lower(t."name");`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx, getTopicsQuery)
  if nil != err {
    slog.Error(getErrMsg(err))
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
      slog.Error(getErrMsg(err))
      return nil, err
    }

    topics = append(topics, &topic)
  }

  return topics, nil
}

// Update updates an existing topic.
func (r *TopicsRepository) Update(ctx context.Context, id string, update *transfer.TopicUpdate) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  updateArticleTopicQuery := `
  UPDATE "archive"."article"
     SET "topic" = $2
   WHERE "topic" = $1;`

  ctx1, cancel := context.WithTimeout(ctx, 7*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx1, updateArticleTopicQuery, id, update.ID)

  if nil != err {
    if strings.Contains(err.Error(), `violates foreign key constraint "article_topic_fkey"`) {
      p := &problem.Problem{}
      p.Type(problem.TypeActionRefused)
      p.Status(http.StatusBadRequest)
      p.Title("Could not update topic.")
      p.Detail("Cannot update topic name because it is already in used by some articles. Try deleting this one and register a new topic under this name.")
      p.With("name", update.Name)
      return p
    }
    slog.Error(getErrMsg(err))
    return err
  }

  updateTopicQuery := `
  UPDATE "archive"."topic"
     SET "id" = $2,
         "name" = $3,
         "updated_at" = current_timestamp
   WHERE "id" = $1;`

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err = tx.ExecContext(ctx, updateTopicQuery,
    id,
    update.ID,
    update.Name,
  )

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "topic")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// Remove removes a topic and detaches it from any article that uses it.
func (r *TopicsRepository) Remove(ctx context.Context, id string) error {
  slog.Info("removing article topic", slog.String("id", id))

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  removeTopicQuery := `
  DELETE FROM "archive"."topic"
        WHERE "id" = $1;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx1, removeTopicQuery, id)

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "topic")
  }

  removeFromAttachedArticlesQuery := `
  UPDATE "archive"."article"
     SET "topic" = NULL
   WHERE "topic" = $1;`

  ctx1, cancel = context.WithTimeout(ctx, 7*time.Second)
  defer cancel()

  result, err = tx.ExecContext(ctx1, removeFromAttachedArticlesQuery, id)

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if _, err = result.RowsAffected(); nil != err {
    slog.Error(getErrMsg(err))
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

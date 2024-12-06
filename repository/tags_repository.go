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

// TagsRepository is a low level API that provides methods for interacting
// with tags in the database.
type TagsRepository struct {
  db *sql.DB
}

func NewTagsRepository(db *sql.DB) *TagsRepository {
  return &TagsRepository{db}
}

// Create adds a new tag.
func (r *TagsRepository) Create(ctx context.Context, creation *transfer.TagCreation) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    return err
  }

  defer tx.Rollback()

  addTagQuery := `
  INSERT INTO "archive"."tag" ("id", "name")
               VALUES ($1, $2);`

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, addTagQuery,
    creation.ID,
    creation.Name,
  )

  if nil != err {
    if strings.Contains(err.Error(), `duplicate key value violates unique constraint "tag_pkey"`) ||
      strings.Contains(err.Error(), `duplicate key value violates unique constraint "tag_name_key"`) {
      var p problem.Problem
      p.Type(problem.TypeDuplicateKey)
      p.Status(http.StatusConflict)
      p.Title("Duplicate tag.")
      p.Detail("A tag with a similar name is already registered. Try using a different one.")
      p.With("name", creation.Name)
      return &p
    }
    slog.Error(getErrMsg(err))
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    p := problem.Problem{}
    p.Title("Tag not created.")
    p.Detail("Could not create tag for an unknown reason.")
    p.Status(http.StatusInternalServerError)

    return &p
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// List retrieves all the tags.
func (r *TagsRepository) List(ctx context.Context) (tags []*model.Tag, err error) {
  getTagsQuery := `
  SELECT t."id",
         t."name",
         t."created_at",
         t."updated_at"
    FROM "archive"."tag" t
   WHERE 1 <= (SELECT count(a.*)
                 FROM "archive"."article" a
           INNER JOIN "archive"."article_tag" at ON at."article_uuid" = a."uuid"
                WHERE NOT a."hidden"
                      AND a."published_at" IS NOT NULL
                      AND at."tag_id" = t."id")
ORDER BY lower(t."name");`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx, getTagsQuery)
  if nil != err {
    slog.Error(getErrMsg(err))
    return nil, err
  }

  defer result.Close()

  tags = make([]*model.Tag, 0)

  for result.Next() {
    var tag model.Tag

    err = result.Scan(
      &tag.ID,
      &tag.Name,
      &tag.CreatedAt,
      &tag.UpdatedAt,
    )

    if nil != err {
      slog.Error(getErrMsg(err))
      return nil, err
    }

    tags = append(tags, &tag)
  }

  return tags, nil
}

// Update updates an existing tag.
func (r *TagsRepository) Update(ctx context.Context, id string, update *transfer.TagUpdate) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  updateArticleTagQuery := `
  UPDATE "archive"."article_tag"
     SET "tag_id" = $2
   WHERE "tag_id" = $1;`

  ctx1, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx1, updateArticleTagQuery,
    id,
    update.ID,
  )

  if nil != err {
    if strings.Contains(err.Error(), `violates foreign key constraint "article_tag_tag_id_fkey"`) {
      p := &problem.Problem{}
      p.Type(problem.TypeActionRefused)
      p.Status(http.StatusBadRequest)
      p.Title("Could not update article tag.")
      p.Detail("Cannot update tag name because it is already in used by some articles. Try deleting this one and register a new tag under this name.")
      p.With("name", update.Name)
      return p
    }
    slog.Error(getErrMsg(err))
    return err
  }

  updateTagQuery := `
  UPDATE "archive"."tag"
     SET "id" = $2,
         "name" = $3,
         "updated_at" = current_timestamp
   WHERE "id" = $1;`

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err = tx.ExecContext(ctx, updateTagQuery,
    id,
    update.ID,
    update.Name,
  )

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "tag")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// Remove removes a tag and detaches it from any article that uses it.
func (r *TagsRepository) Remove(ctx context.Context, id string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  removeTagQuery := `
  DELETE FROM "archive"."tag"
        WHERE "id" = $1;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx1, removeTagQuery, id)

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "tag")
  }

  removeFromAttachedArticlesQuery := `
  DELETE FROM "archive"."article_tag"
        WHERE "tag_id" = $1;`

  ctx1, cancel = context.WithTimeout(ctx, 5*time.Second)
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

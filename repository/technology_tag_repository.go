package repository

import (
  "context"
  "database/sql"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "log/slog"
  "time"
)

// TechnologyTagRepository provides methods for interacting with technology tags data in the database.
type TechnologyTagRepository struct {
  db *sql.DB
}

func NewTechnologyTagRepository(db *sql.DB) *TechnologyTagRepository {
  return &TechnologyTagRepository{db}
}

// List retrieves a slice of technology tags.
func (r *TechnologyTagRepository) List(ctx context.Context) (technologies []*model.TechnologyTag, err error) {
  var getTagsQuery = `
  SELECT *
    FROM "projects"."tag"
ORDER BY "created_at" DESC;`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  rows, err := r.db.QueryContext(ctx, getTagsQuery)
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }
  technologies = make([]*model.TechnologyTag, 0)
  for rows.Next() {
    var tech = new(model.TechnologyTag)
    err = rows.Scan(&tech.UUID, &tech.Name, &tech.CreatedAt, &tech.UpdatedAt)
    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }
    technologies = append(technologies, tech)
  }
  return
}

// Create creates a new technology tag record with the provided creation data.
func (r *TechnologyTagRepository) Create(ctx context.Context, creation *transfer.TechnologyTagCreation) (id string, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }
  defer tx.Rollback()
  var addTagQuery = `
  INSERT INTO "projects"."tag" ("name") VALUES ($1)
    RETURNING "uuid";`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  var result = tx.QueryRowContext(ctx, addTagQuery, creation.Name)
  err = result.Scan(&id)
  if nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }
  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }
  return id, nil
}

// Exists checks whether a technology tag exists in the database.
// If it does, it returns nil; otherwise a not found error.
func (r *TechnologyTagRepository) Exists(ctx context.Context, id string) (err error) {
  var existsTagQuery = `
  SELECT count (1)
    FROM "projects"."tag"
   WHERE "uuid" = $1;`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  var row = r.db.QueryRowContext(ctx, existsTagQuery, id)
  var count int
  err = row.Scan(&count)
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  if count != 1 {
    return problem.NewNotFound(id, "technology_tag")
  }
  return nil
}

// Update modifies an existing technology tag record with the provided update data.
func (r *TechnologyTagRepository) Update(ctx context.Context, id string, update *transfer.TechnologyTagUpdate) (updated bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  defer tx.Rollback()
  var updateTagQuery = `
  SELECT "name"
    FROM "projects"."tag"
   WHERE "uuid" = $1;`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  var row = tx.QueryRowContext(ctx, updateTagQuery, id)
  var currentName string
  err = row.Scan(&currentName)
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  if update.Name == currentName {
    return false, nil
  }
  updateTagQuery = `
  UPDATE "projects"."tag"
     SET "name" = $2,
         "updated_at" = current_timestamp
   WHERE "uuid" = $1;`
  ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, updateTagQuery, id, update.Name)
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  affected, _ := result.RowsAffected()
  if 1 != affected {
    return false, nil
  }
  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return false, err
  }
  return true, nil
}

// Remove deletes an existing technology tag. If not found, returns a not found error.
func (r *TechnologyTagRepository) Remove(ctx context.Context, id string) (err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  defer tx.Rollback()
  var removeTagQuery = `
  DELETE FROM "projects"."tag"
        WHERE "uuid" = $1;`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, removeTagQuery, id)
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "technology_tag")
  }
  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }
  return nil
}

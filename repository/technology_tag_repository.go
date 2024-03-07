package repository

import (
  "context"
  "database/sql"
  "fontseca/model"
  "fontseca/problem"
  "fontseca/transfer"
  "log/slog"
  "time"
)

// TechnologyTagRepository provides methods for interacting with technology tags data in the database.
type TechnologyTagRepository interface {
  // Get retrieves a slice of technology tags.
  Get(ctx context.Context) (technologies []*model.TechnologyTag, err error)

  // Add creates a new technology tag record with the provided creation data.
  Add(ctx context.Context, creation *transfer.TechnologyTagCreation) (id string, err error)

  // Exists checks whether a technology tag exists in the database.
  // If it does, it returns nil; otherwise a not found error.
  Exists(ctx context.Context, id string) (err error)

  // Update modifies an existing technology tag record with the provided update data.
  Update(ctx context.Context, id string, update *transfer.TechnologyTagUpdate) (updated bool, err error)

  // Remove deletes an existing technology tag. If not found, returns a not found error.
  Remove(ctx context.Context, id string) (err error)
}

type technologyTagRepository struct {
  db *sql.DB
}

func NewTechnologyTagRepository(db *sql.DB) TechnologyTagRepository {
  return &technologyTagRepository{db}
}

func (r *technologyTagRepository) Get(ctx context.Context) (technologies []*model.TechnologyTag, err error) {
  var query = `SELECT * FROM "technology_tag";`
  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()
  rows, err := r.db.QueryContext(ctx, query)
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }
  technologies = make([]*model.TechnologyTag, 0)
  for rows.Next() {
    var tech = new(model.TechnologyTag)
    err = rows.Scan(&tech.ID, &tech.Name, &tech.CreatedAt, &tech.UpdatedAt)
    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }
    technologies = append(technologies, tech)
  }
  return
}

func (r *technologyTagRepository) Add(ctx context.Context, creation *transfer.TechnologyTagCreation) (id string, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return "", err
  }
  defer tx.Rollback()
  var query = `
  INSERT INTO "technology_tag" ("name")
                        VALUES (@name)
    RETURNING "id";`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var result = tx.QueryRowContext(ctx, query, sql.Named("name", creation.Name))
  err = result.Scan(&id)
  if nil != err {
    slog.Error(err.Error())
    return "", err
  }
  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return "", err
  }
  return id, nil
}

func (r *technologyTagRepository) Exists(ctx context.Context, id string) (err error) {
  var query = `
  SELECT count (1)
    FROM "technology_tag"
   WHERE "id" = @id;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var row = r.db.QueryRowContext(ctx, query, sql.Named("id", id))
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

func (r *technologyTagRepository) Update(ctx context.Context, id string, update *transfer.TechnologyTagUpdate) (updated bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  defer tx.Rollback()
  var query = `
  SELECT "name"
    FROM "technology_tag"
   WHERE id = @id;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var row = tx.QueryRowContext(ctx, query, sql.Named("id", id))
  var currentName string
  err = row.Scan(&currentName)
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  query = `
  UPDATE "technology_tag"
     SET "name" = coalesce (nullif (@new_name, ''), @current_name),
         "updated_at" = current_timestamp
   WHERE "id" = @id;`
  ctx, cancel = context.WithTimeout(ctx, time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, query,
    sql.Named("id", id),
    sql.Named("new_name", update.Name),
    sql.Named("current_name", currentName))
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

func (r *technologyTagRepository) Remove(ctx context.Context, id string) (err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  defer tx.Rollback()
  var query = `
  DELETE FROM "technology_tag"
        WHERE id = @id;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, query, sql.Named("id", id))
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

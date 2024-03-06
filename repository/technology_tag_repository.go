package repository

import (
  "context"
  "database/sql"
  "fontseca/model"
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
    return "", nil
  }
  return id, nil
}

func (r *technologyTagRepository) Update(ctx context.Context, id string, update *transfer.TechnologyTagUpdate) (updated bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *technologyTagRepository) Remove(ctx context.Context, id string) (err error) {
  // TODO implement me
  panic("implement me")
}

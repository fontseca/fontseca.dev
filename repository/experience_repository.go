package repository

import (
  "context"
  "database/sql"
  "fontseca/model"
  "fontseca/transfer"
  "log/slog"
  "time"
)

// ExperienceRepository provides methods for interacting with experience data in the database.
type ExperienceRepository interface {
  // Get retrieves a slice of experience. If hidden is true it returns all
  // the hidden experience records.
  Get(ctx context.Context, hidden ...bool) (experience []*model.Experience, err error)

  // GetByID retrieves a single experience record by its ID.
  GetByID(ctx context.Context, id string) (experience *model.Experience, err error)

  // Save creates a new experience record with the provided creation data.
  Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error)

  // Update modifies an existing experience record with the provided update data.
  Update(ctx context.Context, update *transfer.ExperienceUpdate) (updated bool, err error)

  // Remove deletes an experience record by its ID.
  Remove(ctx context.Context, id string) error
}

type experienceRepository struct {
  db *sql.DB
}

// NewExperienceRepository creates a new ExperienceRepository instance associating db as its database.
func NewExperienceRepository(db *sql.DB) ExperienceRepository {
  return &experienceRepository{db}
}

func (r *experienceRepository) Get(ctx context.Context, hidden ...bool) (experience []*model.Experience, err error) {
  var query string
  if 0 < len(hidden) && hidden[0] {
    query = `SELECT *
               FROM "experience"
              WHERE "hidden" IS TRUE;`
  } else {
    query = `SELECT *
               FROM "experience";`
  }
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  rows, err := r.db.QueryContext(ctx, query)
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }
  s := make([]*model.Experience, 0)
  for rows.Next() {
    e := new(model.Experience)
    err = rows.Scan(
      &e.ID,
      &e.Starts,
      &e.Ends,
      &e.JobTitle,
      &e.Company,
      &e.Country,
      &e.Summary,
      &e.Active,
      &e.Hidden,
      &e.CreatedAt,
      &e.UpdatedAt)
    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }
    s = append(s, e)
  }
  return s, nil
}

func (r *experienceRepository) GetByID(ctx context.Context, id string) (experience *model.Experience, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *experienceRepository) Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  defer tx.Rollback()
  query := `
  INSERT INTO "experience" ("starts",
                            "ends",
                            "job_title",
                            "company",
                            "country",
                            "summary")
                    VALUES (@starts,
                            @ends,
                            @job_title,
                            @company,
                            @country,
                            @summary);`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  result, err := r.db.ExecContext(ctx, query,
    sql.Named("starts", creation.Starts),
    sql.Named("ends", creation.Ends),
    sql.Named("job_title", creation.JobTitle),
    sql.Named("company", creation.Company),
    sql.Named("country", creation.Country),
    sql.Named("summary", creation.Summary))
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
    return false, nil
  }
  return true, nil
}

func (r *experienceRepository) Update(ctx context.Context, update *transfer.ExperienceUpdate) (updated bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *experienceRepository) Remove(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

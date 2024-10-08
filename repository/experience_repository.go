package repository

import (
  "context"
  "database/sql"
  "errors"
  "fmt"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/lib/pq"
  "log/slog"
  "time"
)

// ExperienceRepository provides methods for interacting with experience data in the database.
type ExperienceRepository struct {
  db *sql.DB
}

// NewExperienceRepository creates a new ExperienceRepository instance associating db as its database.
func NewExperienceRepository(db *sql.DB) *ExperienceRepository {
  return &ExperienceRepository{db}
}

// Get retrieves a slice of experience. If hidden is true it returns all
// the hidden experience records.
func (r *ExperienceRepository) Get(ctx context.Context, hidden bool) (experience []*model.Experience, err error) {
  getMyExperienceQuery := `
  SELECT "uuid",
         "starts",
         "ends",
         "job_title",
         "company",
         "country",
         "summary",
         "active",
         "hidden",
         "created_at",
         "updated_at"
    FROM "me"."experience"
   WHERE "hidden" = $1
ORDER BY "starts" DESC;`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  rows, err := r.db.QueryContext(ctx, getMyExperienceQuery, hidden)

  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  experience = make([]*model.Experience, 0)
  for rows.Next() {
    e := new(model.Experience)
    err = rows.Scan(
      &e.UUID,
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

    experience = append(experience, e)
  }

  return experience, nil
}

func (r *ExperienceRepository) doGetByID(ctx context.Context, id string, strict bool) (experience *model.Experience, err error) {
  getExperienceByIDQuery := `
  SELECT "uuid",
         "starts",
         "ends",
         "job_title",
         "company",
         "country",
         "summary",
         "active",
         "hidden",
         "created_at",
         "updated_at"
    FROM "me"."experience"
   WHERE "uuid" = $1`

  if strict {
    getExperienceByIDQuery += `
    AND hidden IS FALSE;`
  }

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  experience = new(model.Experience)
  err = r.db.QueryRowContext(ctx, getExperienceByIDQuery, id).Scan(
    &experience.UUID,
    &experience.Starts,
    &experience.Ends,
    &experience.JobTitle,
    &experience.Company,
    &experience.Country,
    &experience.Summary,
    &experience.Active,
    &experience.Hidden,
    &experience.CreatedAt,
    &experience.UpdatedAt)

  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      err = problem.NewNotFound(id, "experience")
    } else {
      slog.Error(err.Error())
    }

    return nil, err
  }

  return experience, nil
}

// GetByID retrieves a single experience record by its UUID.
func (r *ExperienceRepository) GetByID(ctx context.Context, id string) (experience *model.Experience, err error) {
  return r.doGetByID(ctx, id, true)
}

// Save creates a new experience record with the provided creation data.
func (r *ExperienceRepository) Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(err.Error())
    return false, err
  }

  defer tx.Rollback()

  saveExperienceQuery := `
  INSERT INTO "me"."experience" ("starts",
                                 "ends",
                                 "job_title",
                                 "company",
                                 "country",
                                 "summary",
                                 "active")
                         VALUES ($1, nullif ($2, 0), $3, $4, $5, $6, TRUE);`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, saveExperienceQuery,
    creation.Starts,
    creation.Ends,
    creation.JobTitle,
    creation.Company,
    creation.Country,
    creation.Summary)

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

func (r *ExperienceRepository) updatable(current *model.Experience, update *transfer.ExperienceUpdate) bool {
  if (0 == update.Starts || update.Starts == current.Starts) &&
    (0 == update.Ends || (nil != current.Ends && update.Ends == *current.Ends)) &&
    ("" == update.JobTitle || update.JobTitle == current.JobTitle) &&
    ("" == update.Company || update.Company == current.Company) &&
    ("" == update.Country || update.Country == current.Country) &&
    ("" == update.Summary || update.Summary == current.Summary) &&
    (update.Active == current.Active) &&
    (update.Hidden == current.Hidden) {
    return false
  }
  return true
}

// Update modifies an existing experience record with the provided update data.
func (r *ExperienceRepository) Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) (updated bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(err.Error())
    return false, err
  }

  current, err := r.doGetByID(ctx, id, false)

  if nil != err {
    return false, err
  }

  if updatable := r.updatable(current, update); !updatable {
    return false, nil
  }

  updateExperienceQuery := `
  UPDATE "me"."experience"
     SET "starts" = CASE WHEN $2::INTEGER = $3::INTEGER OR 0 = $2::INTEGER THEN $3::INTEGER ELSE $2::INTEGER END,
         "ends" = CASE WHEN $4::INTEGER = $5::INTEGER OR 0 = $4::INTEGER THEN $5::INTEGER ELSE $4::INTEGER END,
         "job_title" = coalesce (nullif ($6, ''), $7),
         "company" = coalesce (nullif ($8, ''), $9),
         "country" = coalesce (nullif ($10, ''), $11),
         "summary" = coalesce (nullif ($12, ''), $13),
         "active" = $14,
         "hidden" = $15,
         "updated_at" = current_timestamp
   WHERE "uuid" = $1;`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, updateExperienceQuery,
    id,
    update.Starts, current.Starts,
    update.Ends, current.Ends,
    update.JobTitle, current.JobTitle,
    update.Company, current.Company,
    update.Country, current.Country,
    update.Summary, current.Summary,
    update.Active,
    update.Hidden)

  if nil != err {
    er := &pq.Error{}

    if errors.As(err, &er) {
      fmt.Printf("%#v\n", er)
      return
    }

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

// Remove deletes an experience record by its UUID.
func (r *ExperienceRepository) Remove(ctx context.Context, id string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  removeExperienceQuery := `
  DELETE FROM "me"."experience"
        WHERE "uuid" = $1;`

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, removeExperienceQuery, id)

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  affected, _ := result.RowsAffected()

  if 1 != affected {
    return problem.NewNotFound(id, "experience")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

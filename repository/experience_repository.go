package repository

import (
  "context"
  "database/sql"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
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

// List retrieves a slice of experience. If hidden is true it returns all
// the hidden experience records.
func (r *ExperienceRepository) List(ctx context.Context, hidden bool) (experience []*model.Experience, err error) {
  getMyExperienceQuery := `
  SELECT "uuid",
         "date_start",
         "date_end",
         "job_title",
         "company",
         "company_homepage",
         "country",
         "summary",
         "active",
         "hidden",
         "created_at",
         "updated_at"
    FROM "me"."experience"
   WHERE "hidden" = $1
ORDER BY "date_start" DESC;`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  rows, err := r.db.QueryContext(ctx, getMyExperienceQuery, hidden)

  if nil != err {
    slog.Error(getErrMsg(err))
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
      &e.CompanyHomepage,
      &e.Country,
      &e.Summary,
      &e.Active,
      &e.Hidden,
      &e.CreatedAt,
      &e.UpdatedAt)

    if nil != err {
      slog.Error(getErrMsg(err))
      return nil, err
    }

    experience = append(experience, e)
  }

  return experience, nil
}

func (r *ExperienceRepository) doGet(ctx context.Context, id string, strict bool) (experience *model.Experience, err error) {
  getExperienceByIDQuery := `
  SELECT "uuid",
         "date_start",
         "date_end",
         "job_title",
         "company",
         "company_homepage",
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
    &experience.CompanyHomepage,
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
      slog.Error(getErrMsg(err))
    }

    return nil, err
  }

  return experience, nil
}

// Get retrieves a single experience record by its UUID. If strict is set to true, then it ignores hidden records.
func (r *ExperienceRepository) Get(ctx context.Context, id string, strict bool) (experience *model.Experience, err error) {
  return r.doGet(ctx, id, strict)
}

// Create creates a new experience record with the provided creation data.
func (r *ExperienceRepository) Create(ctx context.Context, creation *transfer.ExperienceCreation) (created string, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(getErrMsg(err))
    return "", err
  }

  defer tx.Rollback()

  saveExperienceQuery := `
  INSERT INTO "me"."experience" ("date_start",
                                 "date_end",
                                 "job_title",
                                 "company",
                                 "company_homepage",
                                 "country",
                                 "summary",
                                 "active",
                                 "hidden")
                         VALUES ($1,
                                 CASE WHEN $2 = '' THEN NULL ELSE $2::DATE END,
                                 $3,
                                 $4,
                                 nullif($5, ''),
                                 $6,
                                 $7,
                                 TRUE,
                                 TRUE)
    RETURNING "uuid";`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  err = tx.QueryRowContext(ctx, saveExperienceQuery,
    creation.Starts,
    creation.Ends,
    creation.JobTitle,
    creation.Company,
    creation.CompanyHomepage,
    creation.Country,
    creation.Summary).Scan(&created)

  if nil != err {
    slog.Error(getErrMsg(err))
    return "", err
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return "", nil
  }

  return created, nil
}

// Update modifies an existing experience record with the provided update data.
func (r *ExperienceRepository) Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  updateExperienceQuery := `
  UPDATE "me"."experience"
     SET "date_start" = CASE WHEN $2 <> '' AND $2::DATE <> "date_start" THEN $2::DATE ELSE "date_start" END,
         "date_end" = CASE WHEN $3 = '2006-01-02' THEN NULL
                           WHEN $3 <> '' AND "date_end" IS NULL OR $3::DATE <> "date_end" THEN $3::DATE
                           ELSE "date_end" END,
         "job_title" = coalesce (nullif ($4, ''), "job_title"),
         "company" = coalesce (nullif ($5, ''), "company"),
         "company_homepage" = coalesce (nullif ($6, ''), "company_homepage"),
         "country" = coalesce (nullif ($7, ''), "country"),
         "summary" = coalesce (nullif ($8, ''), "summary"),
         "updated_at" = current_timestamp
   WHERE "uuid" = $1;`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, updateExperienceQuery,
    id,
    update.Starts,
    update.Ends,
    update.JobTitle,
    update.Company,
    update.CompanyHomepage,
    update.Country,
    update.Summary)

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  affected, _ := result.RowsAffected()

  if 1 != affected {
    return problem.NewNotFound(id, "experience")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

func (r *ExperienceRepository) SetHidden(ctx context.Context, id string, hidden bool) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  query := `
  UPDATE "me"."experience"
     SET "hidden" = $2,
         "updated_at" = current_timestamp
   WHERE "uuid" = $1;`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, query, id, hidden)

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "experience")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// Remove deletes an experience record by its UUID.
func (r *ExperienceRepository) Remove(ctx context.Context, id string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(getErrMsg(err))
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
    slog.Error(getErrMsg(err))
    return err
  }

  affected, _ := result.RowsAffected()

  if 1 != affected {
    return problem.NewNotFound(id, "experience")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

package repository

import (
  "context"
  "database/sql"
  "errors"
  "fontseca/model"
  "fontseca/problem"
  "fontseca/transfer"
  "log/slog"
  "time"
)

// ExperienceRepository provides methods for interacting with experience data in the database.
type ExperienceRepository interface {
  // Get retrieves a slice of experience. If hidden is true it returns all
  // the hidden experience records.
  Get(ctx context.Context, hidden bool) (experience []*model.Experience, err error)

  // GetByID retrieves a single experience record by its ID.
  GetByID(ctx context.Context, id string) (experience *model.Experience, err error)

  // Save creates a new experience record with the provided creation data.
  Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error)

  // Update modifies an existing experience record with the provided update data.
  Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) (updated bool, err error)

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

func (r *experienceRepository) Get(ctx context.Context, hidden bool) (experience []*model.Experience, err error) {
  var query string
  if hidden {
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
  query := `SELECT *
              FROM "experience"
             WHERE "id" = @id;`
  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()
  row := r.db.QueryRowContext(ctx, query, sql.Named("id", id))
  experience = new(model.Experience)
  err = row.Scan(
    &experience.ID,
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

func (r *experienceRepository) updatable(current *model.Experience, update *transfer.ExperienceUpdate) bool {
  if (0 == update.Starts || update.Starts == current.Starts) &&
    (0 == update.Ends || update.Ends == current.Ends) &&
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

func (r *experienceRepository) Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) (updated bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  current, err := r.GetByID(ctx, id)
  if nil != err {
    return false, err
  }
  if updatable := r.updatable(current, update); !updatable {
    return false, nil
  }
  query := `
  UPDATE "experience"
     SET "starts" = CASE WHEN @starts = @current_starts OR 0 = @starts THEN @current_starts ELSE @starts END,
         "ends" = CASE WHEN @ends = @current_ends OR 0 = @ends THEN @current_ends ELSE @ends END,
         "job_title" = coalesce (nullif (@job_title, ''), @current_job_title),
         "company" = coalesce (nullif (@company, ''), @current_company),
         "country" = coalesce (nullif (@country, ''), @current_country),
         "summary" = coalesce (nullif (@summary, ''), @current_summary),
         "active" = @active,
         "hidden" = @hidden,
         "updated_at" = current_timestamp
   WHERE "id" = @id;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, query,
    sql.Named("id", id),
    sql.Named("starts", update.Starts), sql.Named("current_starts", current.Starts),
    sql.Named("ends", update.Ends), sql.Named("current_ends", current.Ends),
    sql.Named("job_title", update.JobTitle), sql.Named("current_job_title", current.JobTitle),
    sql.Named("company", update.Company), sql.Named("current_company", current.Company),
    sql.Named("country", update.Country), sql.Named("current_country", current.Country),
    sql.Named("summary", update.Summary), sql.Named("current_summary", current.Summary),
    sql.Named("active", update.Active),
    sql.Named("hidden", update.Hidden))
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

func (r *experienceRepository) Remove(ctx context.Context, id string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  defer tx.Rollback()
  query := `
  DELETE FROM "experience"
        WHERE "id" = @id;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  result, err := r.db.ExecContext(ctx, query, sql.Named("id", id))
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "experience")
  }
  return nil
}

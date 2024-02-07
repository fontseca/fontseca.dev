package repository

import (
  "context"
  "database/sql"
  "fontseca/model"
  "fontseca/transfer"
  "log/slog"
  "time"
)

// MeRepository is an abstraction of the database that provides
// actions for model.Me management.
type MeRepository interface {
  // Register creates my profile record. It ensures that the record is
  // created only once.
  Register(ctx context.Context)

  // Get retrieves the information of my profile.
  Get(ctx context.Context) (me *model.Me, err error)

  // Update updates the information of my profile.
  Update(ctx context.Context, update transfer.MeUpdate) (ok bool, err error)
}

type meRepositoryImpl struct {
  db *sql.DB
}

// NewRepository creates a new MeRepository instance associating db as its database.
func NewRepository(db *sql.DB) MeRepository {
  return &meRepositoryImpl{db}
}

func (r *meRepositoryImpl) registered(ctx context.Context) bool {
  var n int
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var err = r.db.
    QueryRowContext(ctx, `SELECT count (1) FROM "me";`).
    Scan(&n)
  if nil != err {
    slog.Error(err.Error())
    return false
  }
  return 0 < n
}

func (r *meRepositoryImpl) Register(ctx context.Context) {
  if r.registered(ctx) {
    return
  }
  var query = `
  INSERT INTO "me" ("summary",
                    "email",
                    "company",
                    "location")
            VALUES ('No summary provided.',
                    'email@example.com',
                    'None',
                    'Unknown');`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  _, err := r.db.ExecContext(ctx, query)
  if nil != err {
    slog.Error(err.Error())
  }
}

func (r *meRepositoryImpl) Get(ctx context.Context) (me *model.Me, err error) {
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var row = r.db.QueryRowContext(ctx, `SELECT * FROM "me";`)
  me = new(model.Me)
  err = row.Scan(
    &me.Username,
    &me.FirstName,
    &me.LastName,
    &me.Summary,
    &me.JobTitle,
    &me.Email,
    &me.PhotoURL,
    &me.ResumeURL,
    &me.CodingSince,
    &me.Company,
    &me.Location,
    &me.Hireable,
    &me.GitHubURL,
    &me.LinkedInURL,
    &me.YouTubeURL,
    &me.TwitterURL,
    &me.InstagramURL,
    &me.CreatedAt,
    &me.UpdatedAt)
  if nil != err {
    slog.Error(err.Error())
    return me, err
  }
  return me, nil
}

func (r *meRepositoryImpl) Update(ctx context.Context, update transfer.MeUpdate) (ok bool, err error) {
  // TODO implement me
  panic("implement me")
}

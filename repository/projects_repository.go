package repository

import (
  "context"
  "database/sql"
  "errors"
  "fontseca/model"
  "fontseca/problem"
  "fontseca/transfer"
  "log/slog"
  "strings"
  "time"
)

// ProjectsRepository provides methods for interacting with project data in the database.
type ProjectsRepository interface {
  // Get retrieves a slice of project types.
  Get(ctx context.Context, archived bool) (projects []*model.Project, err error)

  // GetByID retrieves a project type by its ID.
  GetByID(ctx context.Context, id string) (project *model.Project, err error)

  // Add creates a project record with the provided creation data.projectID
  Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error)

  // Update modifies an existing project record with the provided update data.
  Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error)

  // Remove deletes an existing project type. If not found, returns a not found error.
  Remove(ctx context.Context, id string) (err error)

  // AddTechnologyTag adds an existing technology tag that will belong to the project represented by projectID .
  AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error)

  // RemoveTechnologyTag removes a technology tag that belongs to the project represented by projectID.
  RemoveTechnologyTag(ctx context.Context, projectID, technologyID string) (err error)
}

type projectsRepository struct {
  db *sql.DB
}

func NewProjectsRepository(db *sql.DB) ProjectsRepository {
  return &projectsRepository{db}
}

func (r *projectsRepository) Get(ctx context.Context, archived bool) (projects []*model.Project, err error) {
  var query = `
     SELECT p.*, group_concat (tt."name")
       FROM "project" p
  LEFT JOIN "project_technology_tag" ptt
         ON ptt."project_id" = p."id"
  LEFT JOIN "technology_tag" tt
         ON tt."id" = ptt."technology_tag_id"
      WHERE p."archived" IS @archived
   GROUP BY p."id"
   ORDER BY p."created_at" DESC;`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  rows, err := r.db.QueryContext(ctx, query, sql.Named("archived", archived))
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }
  projects = make([]*model.Project, 0)
  for rows.Next() {
    var (
      project = new(model.Project)
      tags    *string
    )
    err = rows.Scan(
      &project.ID,
      &project.Name,
      &project.Homepage,
      &project.Language,
      &project.Summary,
      &project.Content,
      &project.EstimatedTime,
      &project.FirstImageURL,
      &project.SecondImageURL,
      &project.GitHubURL,
      &project.CollectionURL,
      &project.PlaygroundURL,
      &project.Playable,
      &project.Archived,
      &project.Finished,
      &project.CreatedAt,
      &project.UpdatedAt,
      &tags)
    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }
    if nil != tags && "" != *tags {
      project.TechnologyTags = strings.Split(*tags, ",")
    }
    projects = append(projects, project)
  }
  return projects, nil
}

func (r *projectsRepository) GetByID(ctx context.Context, id string) (project *model.Project, err error) {
  var query = `
     SELECT p.*, group_concat (tt."name")
       FROM "project" p
  LEFT JOIN "project_technology_tag" ptt
         ON ptt."project_id" = p."id"
  LEFT JOIN "technology_tag" tt
         ON tt."id" = ptt."technology_tag_id"
      WHERE p."archived" IS FALSE
        AND p."id" = @project_id
   GROUP BY p."id";`
  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()
  var result = r.db.QueryRowContext(ctx, query, sql.Named("project_id", id))
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }
  project = new(model.Project)
  var tags *string
  err = result.Scan(
    &project.ID,
    &project.Name,
    &project.Homepage,
    &project.Language,
    &project.Summary,
    &project.Content,
    &project.EstimatedTime,
    &project.FirstImageURL,
    &project.SecondImageURL,
    &project.GitHubURL,
    &project.CollectionURL,
    &project.PlaygroundURL,
    &project.Playable,
    &project.Archived,
    &project.Finished,
    &project.CreatedAt,
    &project.UpdatedAt,
    &tags)
  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      err = problem.NewNotFound(id, "project")
    } else {
      slog.Error(err.Error())
    }
    return nil, err
  }
  if nil != tags && "" != *tags {
    project.TechnologyTags = strings.Split(*tags, ",")
  }
  return project, nil
}

func (r *projectsRepository) Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return "", err
  }
  defer tx.Rollback()
  var query = `
  INSERT INTO "project" ("name",
                         "homepage",
                         "language",
                         "summary",
                         "content",
                         "estimated_time",
                         "first_image_url",
                         "second_image_url",
                         "github_url",
                         "collection_url",
                         "playground_url",
                         "playable",
                         "archived",
                         "finished")
                 VALUES (@name,
                         nullif (@homepage, ''),
                         nullif (@language, ''),
                         nullif (@summary, ''),
                         nullif (@content, ''),
                         nullif (@estimated_time, 0),
                         nullif (@first_image_url, ''),
                         nullif (@second_image_url, ''),
                         nullif (@github_url, ''),
                         nullif (@collection_url, ''),
                         nullif (@playground_url, ''),
                         @playable,
                         @archived,
                         @finished)
              RETURNING "id";`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var row = tx.QueryRowContext(ctx, query,
    sql.Named("name", creation.Name),
    sql.Named("homepage", creation.Homepage),
    sql.Named("language", creation.Language),
    sql.Named("summary", creation.Summary),
    sql.Named("content", creation.Content),
    sql.Named("estimated_time", creation.EstimatedTime),
    sql.Named("first_image_url", creation.FirstImageURL),
    sql.Named("second_image_url", creation.SecondImageURL),
    sql.Named("github_url", creation.GitHubURL),
    sql.Named("collection_url", creation.CollectionURL),
    sql.Named("playground_url", creation.PlaygroundURL),
    sql.Named("playable", creation.Playable),
    sql.Named("archived", creation.Archived),
    sql.Named("finished", creation.Finished))
  err = row.Scan(&id)
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

func (r *projectsRepository) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *projectsRepository) Remove(ctx context.Context, id string) (err error) {
  // TODO implement me
  panic("implement me")
}

func (r *projectsRepository) AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  defer tx.Rollback()
  var query = `
  INSERT INTO "project_technology_tag" ("project_id",
                                        "technology_tag_id")
                                VALUES (@project_id,
                                        @technology_tag_id);`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, query,
    sql.Named("project_id", projectID),
    sql.Named("technology_tag_id", technologyTagID))
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  affected, _ := result.RowsAffected()
  if 1 != affected {
    return false, nil
  }
  if err = tx.Commit(); nil != err {
    return false, err
  }
  return true, nil
}

func (r *projectsRepository) RemoveTechnologyTag(ctx context.Context, projectID, technologyID string) (err error) {
  // TODO implement me
  panic("implement me")
}

package repository

import (
  "context"
  "database/sql"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
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

  // GetBySlug retrieves a project type by its slug.
  GetBySlug(ctx context.Context, slug string) (project *model.Project, err error)

  // Add creates a project record with the provided creation data.projectID
  Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error)

  // Exists checks whether a project exists in the database.
  // If it does, it returns nil; otherwise a not found error.
  Exists(ctx context.Context, id string) (err error)

  // Update modifies an existing project record with the provided update data.
  Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error)

  // Unarchive makes a project not archived so that it can be normally listed.
  Unarchive(ctx context.Context, id string) (unarchived bool, err error)

  // Remove deletes an existing project type. If not found, returns a not found error.
  Remove(ctx context.Context, id string) (err error)

  // ContainsTechnologyTag checks whether technologyTagID belongs to projectID.
  ContainsTechnologyTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error)

  // AddTechnologyTag adds an existing technology tag that will belong to the project represented by projectID .
  AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error)

  // RemoveTechnologyTag removes a technology tag that belongs to the project represented by projectID.
  RemoveTechnologyTag(ctx context.Context, projectID, technologyTagID string) (removed bool, err error)
}

type projectsRepository struct {
  db *sql.DB
}

func NewProjectsRepository(db *sql.DB) ProjectsRepository {
  return &projectsRepository{db}
}

func (r *projectsRepository) Get(ctx context.Context, archived bool) (projects []*model.Project, err error) {
  var query = `
     SELECT p."id",
            p."name",
            p."slug",
            p."homepage",
            p."language",
            p."summary",
            p."content",
            p."estimated_time",
            p."first_image_url",
            p."second_image_url",
            p."github_url",
            p."collection_url",
            p."playground_url",
            p."playable",
            p."archived",
            p."finished",
            p."created_at",
            p."updated_at",
            group_concat (tt."name")
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
  defer rows.Close()
  projects = make([]*model.Project, 0)
  for rows.Next() {
    var (
      project = new(model.Project)
      tags    *string
    )
    err = rows.Scan(
      &project.ID,
      &project.Name,
      &project.Slug,
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

func (r *projectsRepository) doGetByID(ctx context.Context, id string, ignoreArchived bool) (project *model.Project, err error) {
  var query = `
     SELECT p."id",
            p."name",
            p."slug",
            p."homepage",
            p."language",
            p."summary",
            p."read_time",
            p."content",
            p."estimated_time",
            p."first_image_url",
            p."second_image_url",
            p."github_url",
            p."collection_url",
            p."playground_url",
            p."playable",
            p."archived",
            p."finished",
            p."created_at",
            p."updated_at",
            group_concat (tt."name")
       FROM "project" p
  LEFT JOIN "project_technology_tag" ptt
         ON ptt."project_id" = p."id"
  LEFT JOIN "technology_tag" tt
         ON tt."id" = ptt."technology_tag_id"
      WHERE p."archived" IS @archived
        AND p."id" = @project_id
   GROUP BY p."id";`
  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()
  var result = r.db.QueryRowContext(ctx, query, sql.Named("project_id", id), sql.Named("archived", ignoreArchived))
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }
  project = new(model.Project)
  var tags *string
  err = result.Scan(
    &project.ID,
    &project.Name,
    &project.Slug,
    &project.Homepage,
    &project.Language,
    &project.Summary,
    &project.ReadTime,
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

func (r *projectsRepository) GetByID(ctx context.Context, id string) (project *model.Project, err error) {
  return r.doGetByID(ctx, id, false)
}

func (r *projectsRepository) GetBySlug(ctx context.Context, slug string) (project *model.Project, err error) {
  var query = `
     SELECT p."id",
            p."name",
            p."slug",
            p."homepage",
            p."language",
            p."summary",
            p."read_time",
            p."content",
            p."estimated_time",
            p."first_image_url",
            p."second_image_url",
            p."github_url",
            p."collection_url",
            p."playground_url",
            p."playable",
            p."archived",
            p."finished",
            p."created_at",
            p."updated_at",
            group_concat (tt."name")
       FROM "project" p
  LEFT JOIN "project_technology_tag" ptt
         ON ptt."project_id" = p."id"
  LEFT JOIN "technology_tag" tt
         ON tt."id" = ptt."technology_tag_id"
      WHERE p."archived" IS FALSE
        AND p."slug" = @slug
   GROUP BY p."id"
      LIMIT 1;`
  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()
  var result = r.db.QueryRowContext(ctx, query, sql.Named("slug", slug))
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }
  project = new(model.Project)
  var tags *string
  err = result.Scan(
    &project.ID,
    &project.Name,
    &project.Slug,
    &project.Homepage,
    &project.Language,
    &project.Summary,
    &project.ReadTime,
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
      err = problem.NewSlugNotFound(slug, "project")
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
                         "slug",
                         "homepage",
                         "language",
                         "summary",
                         "read_time",
                         "content",
                         "estimated_time",
                         "first_image_url",
                         "second_image_url",
                         "github_url",
                         "collection_url")
                 VALUES (@name,
                         @slug,
                         nullif (@homepage, ''),
                         nullif (@language, ''),
                         nullif (@summary, ''),
                         nullif (@read_time, 0),
                         nullif (@content, ''),
                         nullif (@estimated_time, 0),
                         nullif (@first_image_url, ''),
                         nullif (@second_image_url, ''),
                         nullif (@github_url, ''),
                         nullif (@collection_url, ''))
              RETURNING "id";`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var row = tx.QueryRowContext(ctx, query,
    sql.Named("name", creation.Name),
    sql.Named("slug", creation.Slug),
    sql.Named("homepage", creation.Homepage),
    sql.Named("language", creation.Language),
    sql.Named("summary", creation.Summary),
    sql.Named("content", creation.Content),
    sql.Named("read_time", creation.ReadTime),
    sql.Named("estimated_time", creation.EstimatedTime),
    sql.Named("first_image_url", creation.FirstImageURL),
    sql.Named("second_image_url", creation.SecondImageURL),
    sql.Named("github_url", creation.GitHubURL),
    sql.Named("collection_url", creation.CollectionURL))
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

func (r *projectsRepository) Exists(ctx context.Context, id string) (err error) {
  var query = `
  SELECT count (1)
    FROM "project"
   WHERE "id" = @id;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var row = r.db.QueryRowContext(ctx, query, sql.Named("id", id))
  var exists bool
  err = row.Scan(&exists)
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  if !exists {
    return problem.NewNotFound(id, "project")
  }
  return nil
}

func (r *projectsRepository) nothingToUpdate(current *model.Project, update *transfer.ProjectUpdate) bool {
  return ("" == update.Name || update.Name == current.Name) &&
    ("" == update.Slug || update.Slug == current.Slug) &&
    ("" == update.Homepage || update.Homepage == current.Homepage) &&
    ("" == update.Language || nil != current.Language && update.Language == *current.Language) &&
    ("" == update.Summary || update.Summary == current.Summary) &&
    (0 == update.ReadTime || update.ReadTime == current.ReadTime) &&
    ("" == update.Content || update.Content == current.Content) &&
    (0 == update.EstimatedTime || nil != current.EstimatedTime && update.EstimatedTime == *current.EstimatedTime) &&
    ("" == update.FirstImageURL || update.FirstImageURL == current.FirstImageURL) &&
    ("" == update.SecondImageURL || update.SecondImageURL == current.SecondImageURL) &&
    ("" == update.GitHubURL || update.GitHubURL == current.GitHubURL) &&
    ("" == update.CollectionURL || update.CollectionURL == current.CollectionURL) &&
    ("" == update.PlaygroundURL || update.PlaygroundURL == current.PlaygroundURL) &&
    (update.Archived == current.Archived) &&
    (update.Finished == current.Finished)
}

func (r *projectsRepository) doUpdate(ctx context.Context, id string, update *transfer.ProjectUpdate, ignoreArchived bool) (updated bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  defer tx.Rollback()
  current, err := r.doGetByID(ctx, id, ignoreArchived)
  if nil != err {
    return false, err
  }
  if r.nothingToUpdate(current, update) {
    return false, nil
  }
  var query = `
  UPDATE "project"
     SET "name" = coalesce (nullif (@name, ''), @current_name),
         "slug" = coalesce (nullif (@slug, ''), @current_slug),
         "homepage" = coalesce (nullif (@homepage, ''), @current_homepage),
         "language" = coalesce (nullif (@language, ''), @current_language),
         "summary" = coalesce (nullif (@summary, ''), @current_summary),
         "read_time" = CASE WHEN @read_time = @current_read_time
                              OR 0 = @read_time
                            THEN @current_read_time
                            ELSE @read_time
                            END,
         "content" = coalesce (nullif (@content, ''), @current_content),
         "estimated_time" = CASE WHEN @estimated_time = @current_estimated_time
                                   OR 0 = @estimated_time
                                 THEN @current_estimated_time
                                 ELSE @estimated_time
                                  END,
         "first_image_url" = coalesce (nullif (@first_image_url, ''), @current_first_image_url),
         "second_image_url" = coalesce (nullif (@second_image_url, ''), @current_second_image_url),
         "github_url" = coalesce (nullif (@github_url, ''), @current_github_url),
         "collection_url" = coalesce (nullif (@collection_url, ''), @current_collection_url),
         "playground_url" = coalesce (nullif (@playground_url, ''), @current_playground_url),
         "playable" = @playable,
         "archived" = @archived,
         "finished" = @finished,
         "updated_at" = current_timestamp
   WHERE "id" = @id;`
  var playable = current.Playable
  if "" != update.PlaygroundURL {
    var wantsToDefaultPlaygroundURL = "about:blank" == update.PlaygroundURL
    var notSamePlaygroundURLs = update.PlaygroundURL != current.PlaygroundURL
    if playable && wantsToDefaultPlaygroundURL && notSamePlaygroundURLs {
      playable = false
    } else if !playable && notSamePlaygroundURLs {
      playable = true
    }
  }
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, query, sql.Named("id", id),
    sql.Named("name", update.Name), sql.Named("current_name", current.Name),
    sql.Named("slug", update.Slug), sql.Named("current_slug", current.Slug),
    sql.Named("homepage", update.Homepage), sql.Named("current_homepage", current.Homepage),
    sql.Named("language", update.Language), sql.Named("current_language", current.Language),
    sql.Named("summary", update.Summary), sql.Named("current_summary", current.Summary),
    sql.Named("read_time", update.ReadTime), sql.Named("current_read_time", current.ReadTime),
    sql.Named("content", update.Content), sql.Named("current_content", current.Content),
    sql.Named("estimated_time", update.EstimatedTime), sql.Named("current_estimated_time", current.EstimatedTime),
    sql.Named("first_image_url", update.FirstImageURL), sql.Named("current_first_image_url", current.FirstImageURL),
    sql.Named("second_image_url", update.SecondImageURL), sql.Named("current_second_image_url", current.SecondImageURL),
    sql.Named("github_url", update.GitHubURL), sql.Named("current_github_url", current.GitHubURL),
    sql.Named("collection_url", update.CollectionURL), sql.Named("current_collection_url", current.CollectionURL),
    sql.Named("playground_url", update.PlaygroundURL), sql.Named("current_playground_url", current.PlaygroundURL),
    sql.Named("playable", playable),
    sql.Named("archived", update.Archived),
    sql.Named("finished", update.Finished))
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  var affected, _ = result.RowsAffected()
  if 1 != affected {
    return false, nil
  }
  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return false, err
  }
  return true, nil
}

func (r *projectsRepository) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error) {
  return r.doUpdate(ctx, id, update, false)
}

func (r *projectsRepository) Unarchive(ctx context.Context, id string) (updated bool, err error) {
  return r.doUpdate(ctx, id, &transfer.ProjectUpdate{Archived: false}, true)
}

func (r *projectsRepository) Remove(ctx context.Context, id string) (err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  defer tx.Rollback()

  // Remove technology tags associated with the project to remove.

  var query = `
  DELETE FROM "project_technology_tag"
        WHERE "project_id" = @project_id`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  _, err = tx.ExecContext(ctx, query, sql.Named("project_id", id))
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  // Remove the actual project.

  query = `
  DELETE FROM "project"
        WHERE "id" = @id;`
  ctx, cancel = context.WithTimeout(ctx, time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, query, sql.Named("id", id))
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "project")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *projectsRepository) ContainsTechnologyTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error) {
  var query = `
  SELECT count (1)
    FROM "project_technology_tag" ptt
   WHERE ptt."project_id" = @project_id
     AND ptt."technology_tag_id" = @technology_tag_id;`
  var result = r.db.QueryRowContext(ctx, query,
    sql.Named("project_id", projectID), sql.Named("technology_tag_id", technologyTagID))
  err = result.Scan(&success)
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  return success, nil
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

func (r *projectsRepository) RemoveTechnologyTag(ctx context.Context, projectID, technologyTagID string) (removed bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  defer tx.Rollback()
  var query = `
  DELETE FROM "project_technology_tag"
        WHERE "project_id" = @project_id
          AND "technology_tag_id" = @technology_tag_id;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, query,
    sql.Named("project_id", projectID), sql.Named("technology_tag_id", technologyTagID))
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

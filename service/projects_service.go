package service

import (
  "context"
  "errors"
  "fontseca/model"
  "fontseca/problem"
  "fontseca/repository"
  "fontseca/transfer"
  "log/slog"
  "net/http"
  "regexp"
  "strings"
)

// ProjectsService provides methods for interacting with projects
// data at a higher level and allows extra validation.
type ProjectsService interface {
  // Get retrieves a slice of projects.
  Get(ctx context.Context, archived ...bool) (projects []*model.Project, err error)

  // GetByID retrieves a single project by its ID.
  GetByID(ctx context.Context, id string) (project *model.Project, err error)

  // GetBySlug retrieves a project type by its slug.
  GetBySlug(ctx context.Context, slug string) (project *model.Project, err error)

  // Add creates a project record with the provided creation data.
  Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error)

  // Exists checks whether a given project exists in the database.
  // If it does, it returns nil; otherwise a not found error.
  Exists(ctx context.Context, id string) (err error)

  // Update modifies an existing project record with the provided update data.
  Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error)

  // Unarchive makes a project not archived so that it can be normally listed.
  Unarchive(ctx context.Context, id string) (unarchived bool, err error)

  // Remove deletes an existing project. If not found, returns a not found error.
  Remove(ctx context.Context, id string) (err error)

  // ContainsTechnologyTag checks whether technologyTagID belongs to projectID.
  ContainsTechnologyTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error)

  // AddTechnologyTag adds an existing technology tag that will belong to the project represented by projectID .
  AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error)

  // RemoveTechnologyTag removes a technology tag that belongs to the project represented by projectID.
  RemoveTechnologyTag(ctx context.Context, projectID, technologyTagID string) (removed bool, err error)
}

type projectsService struct {
  r repository.ProjectsRepository
}

func NewProjectsService(repository repository.ProjectsRepository) ProjectsService {
  return &projectsService{repository}
}

func (s *projectsService) Get(ctx context.Context, archived ...bool) (projects []*model.Project, err error) {
  var a = false
  if 0 != len(archived) && archived[0] {
    a = true
  }
  return s.r.Get(ctx, a)
}

func (s *projectsService) GetByID(ctx context.Context, id string) (project *model.Project, err error) {
  if err = validateUUID(&id); err != nil {
    return nil, err
  }
  return s.r.GetByID(ctx, id)
}

func (s *projectsService) GetBySlug(ctx context.Context, slug string) (project *model.Project, err error) {
  return s.r.GetBySlug(ctx, slug)
}

func (s *projectsService) Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error) {
  if nil == creation {
    err = errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return "", err
  }

  removeSpaceInterceptions, err := regexp.Compile(`\s+`)
  if nil != err {
    slog.Error(err.Error())
    return "", err
  }

  creation.Name = strings.TrimSpace(creation.Name)
  creation.Name = removeSpaceInterceptions.ReplaceAllString(creation.Name, " ")
  creation.Homepage = strings.TrimSpace(creation.Homepage)
  creation.Language = strings.TrimSpace(creation.Language)
  creation.Summary = strings.TrimSpace(creation.Summary)
  creation.Content = strings.TrimSpace(creation.Content)
  creation.FirstImageURL = strings.TrimSpace(creation.FirstImageURL)
  creation.SecondImageURL = strings.TrimSpace(creation.SecondImageURL)
  creation.GitHubURL = strings.TrimSpace(creation.GitHubURL)
  creation.CollectionURL = strings.TrimSpace(creation.CollectionURL)

  switch {
  case 0 != len(creation.Name) && 36 < len(creation.Name):
    return "", problem.NewValidation([3]string{"name", "max", "36"})
  case 0 != len(creation.Homepage) && 2048 < len(creation.Homepage):
    return "", problem.NewValidation([3]string{"homepage", "max", "2048"})
  case 0 != len(creation.Language) && 64 < len(creation.Language):
    return "", problem.NewValidation([3]string{"language", "max", "64"})
  case 0 != len(creation.Summary) && 1024 < len(creation.Summary):
    return "", problem.NewValidation([3]string{"summary", "max", "1024"})
  case 0 != len(creation.Content) && 3145728 < len(creation.Content):
    return "", problem.NewValidation([3]string{"content", "max", "3145728"})
  case 0 != len(creation.FirstImageURL) && 2048 < len(creation.FirstImageURL):
    return "", problem.NewValidation([3]string{"first_image_url", "max", "2048"})
  case 0 != len(creation.SecondImageURL) && 2048 < len(creation.SecondImageURL):
    return "", problem.NewValidation([3]string{"second_image_url", "max", "2048"})
  case 0 != len(creation.GitHubURL) && 2048 < len(creation.GitHubURL):
    return "", problem.NewValidation([3]string{"github_url", "max", "2048"})
  case 0 != len(creation.CollectionURL) && 2048 < len(creation.CollectionURL):
    return "", problem.NewValidation([3]string{"collection_url", "max", "2048"})
  }

  var projectText strings.Builder

  if "" != creation.Name {
    projectText.WriteString(creation.Name)
  }

  if "" != creation.Summary {
    if 0 < projectText.Len() {
      projectText.WriteRune('\n')
    }
    projectText.WriteString(creation.Summary)
  }

  if "" != creation.Content {
    if 0 < projectText.Len() {
      projectText.WriteRune('\n')
    }
    projectText.WriteString(creation.Content)
  }

  if 0 < projectText.Len() {
    var r = strings.NewReader(projectText.String())
    creation.ReadTime = computePostReadingTimeInMinutes(r)
  }

  creation.Slug = strings.ToLower(strings.ReplaceAll(creation.Name, " ", "-"))
  err = sanitizeURL(
    &creation.Homepage,
    &creation.FirstImageURL,
    &creation.SecondImageURL,
    &creation.GitHubURL,
    &creation.CollectionURL)
  if nil != err {
    return "", err
  }

  return s.r.Add(ctx, creation)
}

func (s *projectsService) Exists(ctx context.Context, id string) (err error) {
  if err = validateUUID(&id); err != nil {
    return err
  }
  return s.r.Exists(ctx, id)
}

func (s *projectsService) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error) {
  if nil == update {
    err = errors.New("nil value for parameter: update")
    slog.Error(err.Error())
    return false, err
  }

  if err = validateUUID(&id); err != nil {
    return false, err
  }

  removeSpaceInterceptions, err := regexp.Compile(`\s+`)
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }

  update.Name = strings.TrimSpace(update.Name)

  if "" != update.Name {
    update.Name = removeSpaceInterceptions.ReplaceAllString(update.Name, " ")
  }

  update.Homepage = strings.TrimSpace(update.Homepage)
  update.Language = strings.TrimSpace(update.Language)
  update.Summary = strings.TrimSpace(update.Summary)
  update.Content = strings.TrimSpace(update.Content)
  update.FirstImageURL = strings.TrimSpace(update.FirstImageURL)
  update.SecondImageURL = strings.TrimSpace(update.SecondImageURL)
  update.GitHubURL = strings.TrimSpace(update.GitHubURL)
  update.CollectionURL = strings.TrimSpace(update.CollectionURL)
  update.PlaygroundURL = strings.TrimSpace(update.PlaygroundURL)

  switch {
  case 0 != len(update.Name) && 36 < len(update.Name):
    return false, problem.NewValidation([3]string{"name", "max", "36"})
  case 0 != len(update.Homepage) && 2048 < len(update.Homepage):
    return false, problem.NewValidation([3]string{"homepage", "max", "2048"})
  case 0 != len(update.Language) && 64 < len(update.Language):
    return false, problem.NewValidation([3]string{"language", "max", "64"})
  case 0 != len(update.Summary) && 1024 < len(update.Summary):
    return false, problem.NewValidation([3]string{"summary", "max", "1024"})
  case 0 != len(update.Content) && 3145728 < len(update.Content):
    return false, problem.NewValidation([3]string{"content", "max", "3145728"})
  case 0 != len(update.FirstImageURL) && 2048 < len(update.FirstImageURL):
    return false, problem.NewValidation([3]string{"first_image_url", "max", "2048"})
  case 0 != len(update.SecondImageURL) && 2048 < len(update.SecondImageURL):
    return false, problem.NewValidation([3]string{"second_image_url", "max", "2048"})
  case 0 != len(update.GitHubURL) && 2048 < len(update.GitHubURL):
    return false, problem.NewValidation([3]string{"github_url", "max", "2048"})
  case 0 != len(update.CollectionURL) && 2048 < len(update.CollectionURL):
    return false, problem.NewValidation([3]string{"collection_url", "max", "2048"})
  case 0 != len(update.PlaygroundURL) && 2048 < len(update.PlaygroundURL):
    return false, problem.NewValidation([3]string{"playground_url", "max", "2048"})
  }

  var projectText strings.Builder

  if "" != update.Name {
    projectText.WriteString(update.Name)
  }

  if "" != update.Summary {
    if 0 < projectText.Len() {
      projectText.WriteRune('\n')
    }
    projectText.WriteString(update.Summary)
  }

  if "" != update.Content {
    if 0 < projectText.Len() {
      projectText.WriteRune('\n')
    }
    projectText.WriteString(update.Content)
  }

  if 0 < projectText.Len() {
    var r = strings.NewReader(projectText.String())
    update.ReadTime = computePostReadingTimeInMinutes(r)
  }

  if "" != update.Name {
    update.Slug = strings.ToLower(removeSpaceInterceptions.ReplaceAllString(update.Name, "-"))
  }

  err = sanitizeURL(
    &update.Homepage,
    &update.FirstImageURL,
    &update.SecondImageURL,
    &update.GitHubURL,
    &update.CollectionURL,
    &update.PlaygroundURL)
  if nil != err {
    return false, err
  }

  return s.r.Update(ctx, id, update)
}

func (s *projectsService) Unarchive(ctx context.Context, id string) (unarchived bool, err error) {
  if err = validateUUID(&id); err != nil {
    return false, err
  }
  return s.r.Unarchive(ctx, id)
}

func (s *projectsService) Remove(ctx context.Context, id string) (err error) {
  if err = validateUUID(&id); err != nil {
    return err
  }
  return s.r.Remove(ctx, id)
}

func (s *projectsService) ContainsTechnologyTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error) {
  if err = validateUUID(&projectID); err != nil {
    return false, err
  }
  if err = validateUUID(&technologyTagID); err != nil {
    return false, err
  }
  return s.r.ContainsTechnologyTag(ctx, projectID, technologyTagID)
}

func (s *projectsService) AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error) {
  if err = validateUUID(&projectID); err != nil {
    return false, err
  }
  if err = validateUUID(&technologyTagID); err != nil {
    return false, err
  }
  conflict, err := s.ContainsTechnologyTag(ctx, projectID, technologyTagID)
  if nil != err {
    return false, err
  }
  if conflict {
    var p problem.Problem
    p.Status(http.StatusConflict)
    p.Title("Duplicate technology tag.")
    p.Detail("The specified technology tag is already associated with this project. Try using a different one.")
    p.With("project_id", projectID)
    p.With("technology_tag_id", projectID)
    return false, &p
  }
  return s.r.AddTechnologyTag(ctx, projectID, technologyTagID)
}

func (s *projectsService) RemoveTechnologyTag(ctx context.Context, projectID, technologyTagID string) (removed bool, err error) {
  if err = validateUUID(&projectID); err != nil {
    return false, err
  }
  if err = validateUUID(&technologyTagID); err != nil {
    return false, err
  }
  return s.r.RemoveTechnologyTag(ctx, projectID, technologyTagID)
}

package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/repository"
  "fontseca.dev/transfer"
  "log/slog"
  "strconv"
  "strings"
  "time"
)

// ExperienceService provides methods for interacting with
// experience data at a higher level.
type ExperienceService interface {
  // Get returns a slice of experience models and an error if
  // the operation fails. If hidden is true it returns all the
  // hidden experience records.
  Get(ctx context.Context, hidden ...bool) (experience []*model.Experience, err error)

  // GetByID retrieves a single experience record by its UUID.
  GetByID(ctx context.Context, id string) (experience *model.Experience, err error)

  // Save creates a new experience record with the provided creation data.
  // It returns a boolean indicating whether the experience was successfully
  // saved and an error if something went wrong.
  Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error)

  // Update modifies an existing experience record with the provided update data.
  // It returns a boolean indicating whether the experience was successfully updated
  // and an error if something went wrong.
  Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) (updated bool, err error)

  // Remove deletes an experience record by its UUID.
  // It returns an error if the operation fails; for example,
  // if the record does not exist.
  Remove(ctx context.Context, id string) error
}

type experienceService struct {
  r repository.ExperienceRepository
}

func NewExperienceService(r repository.ExperienceRepository) ExperienceService {
  return &experienceService{r}
}

func (s *experienceService) Get(ctx context.Context, hidden ...bool) (experience []*model.Experience, err error) {
  if 0 != len(hidden) && hidden[0] {
    return s.r.Get(ctx, true)
  }
  return s.r.Get(ctx, false)
}

func (s *experienceService) GetByID(ctx context.Context, id string) (experience *model.Experience, err error) {
  if err = validateUUID(&id); nil != err {
    return nil, err
  }
  return s.r.GetByID(ctx, id)
}

func (s *experienceService) Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error) {
  if nil == creation {
    err = errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return false, err
  }
  creation.JobTitle = strings.TrimSpace(creation.JobTitle)
  creation.Company = strings.TrimSpace(creation.Company)
  creation.Country = strings.TrimSpace(creation.Country)
  creation.Summary = strings.TrimSpace(creation.Summary)

  var year = time.Now().Year()
  switch {
  case 0 != creation.Starts && 2017 >= creation.Starts:
    return false, problem.NewValidation([3]string{"starts", "gt", "2017"})
  case 0 != creation.Starts && year < creation.Starts:
    return false, problem.NewValidation([3]string{"starts", "lte", strconv.Itoa(year)})
  case 0 != creation.Ends && creation.Starts > creation.Ends:
    return false, problem.NewValidation([3]string{"ends", "gte", strconv.Itoa(creation.Starts)})
  case 0 != creation.Ends && year < creation.Ends:
    return false, problem.NewValidation([3]string{"ends", "lte", strconv.Itoa(year)})
  }

  return s.r.Save(ctx, creation)
}

func (s *experienceService) Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) (updated bool, err error) {
  if nil == update {
    err = errors.New("nil value for parameter: update")
    slog.Error(err.Error())
    return false, err
  }
  if err = validateUUID(&id); nil != err {
    return false, err
  }

  update.JobTitle = strings.TrimSpace(update.JobTitle)
  update.Company = strings.TrimSpace(update.Company)
  update.Country = strings.TrimSpace(update.Country)
  update.Summary = strings.TrimSpace(update.Summary)

  var year = time.Now().Year()
  switch {
  case 0 != update.Starts && 2017 >= update.Starts:
    return false, problem.NewValidation([3]string{"starts", "gt", "2017"})
  case 0 != update.Starts && year < update.Starts:
    return false, problem.NewValidation([3]string{"starts", "lte", strconv.Itoa(year)})
  case 0 != update.Ends:
    switch {
    case 0 != update.Starts && update.Starts > update.Ends:
      return false, problem.NewValidation([3]string{"ends", "gte", strconv.Itoa(update.Starts)})
    case 2017 >= update.Ends:
      return false, problem.NewValidation([3]string{"ends", "gt", "2017"})
    case year < update.Ends:
      return false, problem.NewValidation([3]string{"ends", "lte", strconv.Itoa(year)})
    }
  }

  return s.r.Update(ctx, id, update)
}

func (s *experienceService) Remove(ctx context.Context, id string) error {
  if err := validateUUID(&id); err != nil {
    return err
  }
  return s.r.Remove(ctx, id)
}

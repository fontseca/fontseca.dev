package service

import (
  "context"
  "errors"
  "fontseca/model"
  "fontseca/repository"
  "fontseca/transfer"
  "log/slog"
  "strings"
)

// MeService defines the interface for managing user profile related operations.
type MeService interface {
  // Get retrieves the information of my profile.
  // It returns client-friendly errors when they occur.
  Get(ctx context.Context) (me *model.Me, err error)

  // Update updates the user profile information with the provided data.
  // It handles validations for the update and returns client-friendly
  // errors when they occur. Returns true if the profile was successfully
  // updated, otherwise false.
  Update(ctx context.Context, update *transfer.MeUpdate) (updated bool, err error)
}

type meService struct {
  r repository.MeRepository
}

func NewMeService(r repository.MeRepository) MeService {
  return &meService{r}
}

func (m *meService) Get(ctx context.Context) (me *model.Me, err error) {
  return m.r.Get(ctx)
}

func (m *meService) Update(ctx context.Context, update *transfer.MeUpdate) (updated bool, err error) {
  if nil == update {
    err = errors.New("nil value for parameter: update")
    slog.Error(err.Error())
    return false, err
  }
  update.Summary = strings.TrimSpace(update.Summary)
  update.JobTitle = strings.TrimSpace(update.JobTitle)
  update.Email = strings.TrimSpace(update.Email)
  update.PhotoURL = strings.TrimSpace(update.PhotoURL)
  update.ResumeURL = strings.TrimSpace(update.ResumeURL)
  update.Company = strings.TrimSpace(update.Company)
  update.Location = strings.TrimSpace(update.Location)
  update.GitHubURL = strings.TrimSpace(update.GitHubURL)
  update.LinkedInURL = strings.TrimSpace(update.LinkedInURL)
  update.YouTubeURL = strings.TrimSpace(update.YouTubeURL)
  update.TwitterURL = strings.TrimSpace(update.TwitterURL)
  update.InstagramURL = strings.TrimSpace(update.InstagramURL)
  return m.r.Update(ctx, update)
}

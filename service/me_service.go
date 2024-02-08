package main

import (
  "context"
  "fontseca/model"
  "fontseca/transfer"
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

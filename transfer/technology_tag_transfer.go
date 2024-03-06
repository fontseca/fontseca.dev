package transfer

// TechnologyTagCreation represents the data required to create a new technology tag entry.
type TechnologyTagCreation struct {
  Name string `json:"name" binding:"required"`
}

// TechnologyTagUpdate represents the data required to update an existing technology tag entry.
type TechnologyTagUpdate struct {
  Name string `json:"name" binding:"required"`
}

package transfer

// TagCreation represents the data required to create a new tag entry.
type TagCreation struct {
  ID   string
  Name string `json:"name" binding:"required,max=32"`
}

// TagUpdate represents the data required to update an existing tag entry.
type TagUpdate struct {
  ID   string
  Name string `json:"name" binding:"required,max=32"`
}

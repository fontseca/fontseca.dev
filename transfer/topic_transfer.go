package transfer

// TopicCreation represents the data required to create a new topic entry.
type TopicCreation struct {
  ID   string
  Name string `json:"name" binding:"required,max=32"`
}

// TopicUpdate represents the data required to update an existing topic entry.
type TopicUpdate struct {
  Name string `json:"name" binding:"required,max=32"`
}

package types

// UserCreationRequest describes the data needed to create a new user
type UserCreationRequest struct {
	Handle   *string `json:"handle"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

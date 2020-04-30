package types

import "github.com/6a/blade-ii-api/pkg/elo"

// UserCreationRequest describes the data needed to create a new user
type UserCreationRequest struct {
	Handle   *string `json:"handle"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

// MMRUpdateRequest describes the data needed to update the MMR for a pair of users after a match
type MMRUpdateRequest struct {
	Player1ID *uint64
	Player2ID *uint64
	Winner    *elo.Player
}

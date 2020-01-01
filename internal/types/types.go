package types

import "encoding/json"

// UserCreationRequest describes the data needed to create a new user
type UserCreationRequest struct {
	Handle   *string `json:"handle"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

// Response is the standard response struct
type Response struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

// Token typedef
type Token byte

// Token Types
const (
	AuthToken Token = iota
	EmailConfirmationToken
	PasswordResetToken
	RefreshToken
)

func (token Token) String() string {
	types := [...]string{
		"auth",
		"email_confirmation",
		"password_reset",
		"refresh",
	}

	if token < AuthToken || token > RefreshToken {
		return "unknown"
	}

	return types[token]
}

// GetAuthTokenResponse is a container for responses to successful get auth requests
type GetAuthTokenResponse struct {
	PublicID     string `json:"pid"`
	AuthToken    string `json:"authToken"`
	RefreshToken string `json:"refreshToken"`
}

// ToJSON returns the reponse as a json format string
func (r GetAuthTokenResponse) ToJSON() (s string) {
	bytes, _ := json.Marshal(r)
	return string(bytes)
}

// UserUpdateRequest is a set of deltas used to update the w/d/l for the specified player
// type UserUpdateRequest struct {
// 	Name   string `json:"name"`
// 	Wins   int    `json:"wins"`
// 	Draws  int    `json:"draws"`
// 	Losses int    `json:"losses"`
// }

// LeaderboardRequest describes the data needed get the leaderboard info , aligned with the specified user
// type LeaderboardRequest struct {
// 	Name string `json:"name"`
// }

// // LeaderboardRow described a single row from a leaderboard
// type LeaderboardRow struct {
// 	Rank   int     `json:"rank"`
// 	OutOf  int     `json:"outof"`
// 	Name   string  `json:"name"`
// 	Wins   int     `json:"wins"`
// 	Ratio  float32 `json:"ratio"`
// 	Draws  int     `json:"draws"`
// 	Losses int     `json:"losses"`
// 	Played int     `json:"played"`
// }

// // Fill will fill a leaderboard row with the provided info
// func (row *LeaderboardRow) Fill(name string, rank int, outof int, wins int, ratio float32, draws int, losses int, played int) {
// 	row.Rank = rank
// 	row.OutOf = outof
// 	row.Name = name
// 	row.Wins = wins
// 	row.Ratio = ratio
// 	row.Draws = draws
// 	row.Losses = losses
// 	row.Played = played
// }

// // Leaderboard describes a leaderboard, as well as an extra container with information for the specified user
// type Leaderboard struct {
// 	User        LeaderboardRow   `json:"user"`
// 	Leaderboard []LeaderboardRow `json:"leaderboard"`
// }

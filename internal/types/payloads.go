// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package types defines types and contstants for this application.
package types

import "time"

// Structs defined here should also include json serialization hints. They are used to wrap return data
// for responses sent back to the client.

// AuthResponsePayload is a container for the response payload of a successful get auth request.
type AuthResponsePayload struct {
	PublicID     string `json:"pid"`
	AuthToken    string `json:"authToken"`
	RefreshToken string `json:"refreshToken"`
}

// CreateAccountResponsePayload is a container for the response payload of a successful account creation request.
type CreateAccountResponsePayload struct {
	Handle string `json:"handle"`
}

// ProfileResponsePayload is a container for the response payload of a successful profile get request.
type ProfileResponsePayload struct {
	Avatar      uint8     `json:"avatar"`
	MMR         int16     `json:"mmr"`
	Wins        uint32    `json:"wins"`
	Draws       uint32    `json:"draws"`
	Losses      uint32    `json:"losses"`
	WinRatio    float32   `json:"winratio"`
	RankedTotal int64     `json:"rankedtotal"`
	Created     time.Time `json:"created"`
}

// LeaderboardResponsePayload is a container for the response payload of a successful leaderboards get request.
type LeaderboardResponsePayload struct {
	Leaderboards []LeaderboardRow `json:"leaderboards"`
	User         LeaderboardRow   `json:"user"`
}

// LeaderboardRow represents a single row in the leaderboards.
type LeaderboardRow struct {
	Handle      string  `json:"handle"`
	Avatar      uint8   `json:"avatar"`
	MMR         int16   `json:"mmr"`
	Wins        uint32  `json:"wins"`
	Draws       uint32  `json:"draws"`
	Losses      uint32  `json:"losses"`
	WinRatio    float32 `json:"winratio"`
	RankedTotal int64   `json:"total"`
	PublicID    string  `json:"pid"`
	Rank        uint64  `json:"rank"`
	OutOf       uint64  `json:"outof"`
}

// MatchHistory is a container for the response payload of a successful match history get request.
type MatchHistory struct {
	Rows []MatchHistoryRow `json:"rows"`
}

// MatchHistoryRow is a single row in a players match history.
type MatchHistoryRow struct {
	MatchID         uint64    `json:"matchid"`
	Player1Handle   string    `json:"player1handle"`
	Player1PublicID string    `json:"player1pid"`
	Player2Handle   string    `json:"player2handle"`
	Player2PublicID string    `json:"player2pid"`
	WinnerHandle    string    `json:"winnerhandle"`
	WinnerPublicID  string    `json:"winnerpid"`
	EndTime         time.Time `json:"endtime"`
}

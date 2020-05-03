package types

import "time"

// AuthResponsePayload is a container for the response payload of a successful get auth request
type AuthResponsePayload struct {
	PublicID     string `json:"pid"`
	AuthToken    string `json:"authToken"`
	RefreshToken string `json:"refreshToken"`
}

// CreateAccountResponsePayload is a container for the response payload of a successful account creation request
type CreateAccountResponsePayload struct {
	Handle string `json:"handle"`
}

// ProfileResponsePayload is a container for the response payload of a successful profile get request
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

// LeaderboardResponsePayload is a container for the response payload of a successful leaderboards get request
type LeaderboardResponsePayload struct {
	Leaderboards []LeaderboardRow `json:"leaderboards"`
	User         LeaderboardRow   `json:"user"`
}

// LeaderboardRow represents a single row in the leaderboards
type LeaderboardRow struct {
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

// Fill populates a leaderboards row object
func (row *LeaderboardRow) Fill(avatar uint8, mmr int16, wins uint32, draws uint32, losses uint32, winratio float32, total int64, publicID string, rank uint64, outOf uint64) {
	row.Avatar = avatar
	row.MMR = mmr
	row.Wins = wins
	row.Draws = draws
	row.Losses = losses
	row.WinRatio = winratio
	row.RankedTotal = total
	row.PublicID = publicID
	row.Rank = rank
	row.OutOf = outOf
}

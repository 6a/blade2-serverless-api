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

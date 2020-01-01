package types

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

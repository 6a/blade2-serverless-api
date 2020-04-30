package elo

// Player is a typedef for the two different possible players (player 1 or player 2)
type Player uint8

// Winner type enums
const (
	Draw Player = iota
	Player1
	Player2
)

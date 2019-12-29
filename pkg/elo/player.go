package elo

// Player is a typedef for the two different possible players (player 1 or player 2)
type Player int

// Winner type enums
const (
	Player1 Player = iota
	Player2 Player = iota
	Draw    Player = iota
)

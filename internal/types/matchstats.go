package types

// MatchStats is a wrapper for a players match stats
type MatchStats struct {
	MMR    int16
	Wins   uint32
	Draws  uint32
	Losses uint32
}

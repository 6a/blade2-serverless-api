// Original formulas and theory: https://www.youtube.com/watch?v=AsYfbmp0To0
//
// Manual testing: https://www.3dkingdoms.com/chess/elo.htm

package elo

import "math"

// Default is the default starting elo for all new users
var Default int = 1200

// Values that are fine as default but can be modified via the associated functions------------------------------------------
var tenXMod float64 = 400.0 // This value is the number of elo points required to be considered 10x better/worse than someone
var k int = 32              // Not sure exactly what k is, but its basically the maximum elo shift after a game

// const values for calculations
const lossScore float64 = 0.0
const drawScore float64 = 0.5
const winScore float64 = 1.0

// CalculateNewElo calculates the new elo ratings for both players depending on the outcome of a match
func CalculateNewElo(player1Elo int16, player2Elo int16, winner Player) (player1NewElo int16, player2NewElo int16) {
	// We dont need the values for player 2 because they can be determined by the values for player 1 instead
	player1WinChance, _ := getWinChance(player1Elo, player2Elo)

	var player1Score float64

	if winner == Player1 {
		player1Score = winScore
	} else if winner == Draw {
		player1Score = drawScore
	} else {
		player1Score = lossScore
	}

	player1EloShift := int16(math.Round(float64(k) * (player1Score - player1WinChance)))
	player2EloShift := player1EloShift * -1

	player1NewElo = player1Elo + player1EloShift
	player2NewElo = player2Elo + player2EloShift

	return player1NewElo, player2NewElo
}

// SetK sets a new internal value for K
func SetK(newK int) {
	k = newK
}

// SetNewTenXMod sets a new internal value for the 10x mod (number of points required to be 10x better/worse than someone
func SetNewTenXMod(newMod float64) {
	tenXMod = newMod
}

// Interal methods ------------------------------------------------------------------------------------------------------------------------

func getWinChance(player1Elo int16, player2Elo int16) (player1Chance float64, player2Chance float64) {
	diff := player2Elo - player1Elo
	diffOver400 := float64(diff) / tenXMod

	player1Chance = 1.0 / (1 + math.Pow(10, diffOver400))
	player2Chance = 1 - player1Chance

	return player1Chance, player2Chance
}

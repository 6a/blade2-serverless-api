// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package elo implements an elo calculator, based on https://www.youtube.com/watch?v=AsYfbmp0To0.
package elo

import "math"

var (

	// Default is the default starting elo for all new users.
	Default int = 1200

	// This value is the number of elo points required to be considered 10x better/worse than someone
	tenXMod float64 = 400.0

	// Not sure exactly what k is, but its basically the maximum elo shift after a game
	k int = 32
)

// const values for elo calculations.
const lossScore float64 = 0.0
const drawScore float64 = 0.5
const winScore float64 = 1.0

// CalculateNewElo calculates the new elo ratings for both players depending on the outcome of a match.
func CalculateNewElo(player1Elo int16, player2Elo int16, winner Player) (player1NewElo int16, player2NewElo int16) {

	// Get the win chance for player 1. We dont need the value for player 2 because player 2's new elo is
	// determined by the values for player 1 instead.
	player1WinChance, _ := getWinChance(player1Elo, player2Elo)

	// Determine the score for player 1 (win draw or loss values).
	var player1Score float64
	if winner == Player1 {
		player1Score = winScore
	} else if winner == Draw {
		player1Score = drawScore
	} else {
		player1Score = lossScore
	}

	// Determine the elo shift for both players.
	eloShift := int16(math.Round(float64(k) * (player1Score - player1WinChance)))

	// Determine the new elo for each player.
	player1NewElo = player1Elo + eloShift
	player2NewElo = player2Elo - eloShift

	// return both new elo values.
	return player1NewElo, player2NewElo
}

// SetK sets a new internal value for K.
func SetK(newK int) {
	k = newK
}

// SetNewTenXMod sets a new internal value for the 10x mod (number of points required to be 10x better/worse than someone.
func SetNewTenXMod(newMod float64) {
	tenXMod = newMod
}

// getWinChance determines the chance for each player to win, based on the elo of both players.
func getWinChance(player1Elo int16, player2Elo int16) (player1Chance float64, player2Chance float64) {

	// Determine the difference in elo.
	diff := player2Elo - player1Elo

	// Divide the difference by the elo difference required to be considered 10 times better than another
	// player.
	diffOverTenX := float64(diff) / tenXMod

	// Determine the win chance for player 1.
	player1Chance = 1.0 / (1 + math.Pow(10, diffOverTenX))

	// Determine the win chance for player 2.
	player2Chance = 1 - player1Chance

	// Return both win chances.
	return player1Chance, player2Chance
}

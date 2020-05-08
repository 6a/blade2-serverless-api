// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package elo implements an elo calculator, based on https://www.youtube.com/watch?v=AsYfbmp0To0.
package elo

import (
	"testing"
)

// Test_CalculateNewElo runs unit tests for the elo calculator.
func Test_CalculateNewElo(t *testing.T) {
	type args struct {
		player1Elo int16
		player2Elo int16
		winner     Player
	}
	tests := []struct {
		name              string
		args              args
		wantPlayer1NewElo int16
		wantPlayer2NewElo int16
	}{
		{
			name: "2k vs 2k player 1 wins",
			args: args{
				player1Elo: 2000,
				player2Elo: 2000,
				winner:     Player1,
			},
			wantPlayer1NewElo: 2016,
			wantPlayer2NewElo: 1984,
		},
		{
			name: "2k vs 2k player 2 wins",
			args: args{
				player1Elo: 2000,
				player2Elo: 2000,
				winner:     Player2,
			},
			wantPlayer1NewElo: 1984,
			wantPlayer2NewElo: 2016,
		},
		{
			name: "2k vs 2k draw",
			args: args{
				player1Elo: 2000,
				player2Elo: 2000,
				winner:     Draw,
			},
			wantPlayer1NewElo: 2000,
			wantPlayer2NewElo: 2000,
		},
		{
			name: "2k vs 2.4k player 1 wins",
			args: args{
				player1Elo: 2000,
				player2Elo: 2400,
				winner:     Player1,
			},
			wantPlayer1NewElo: 2029,
			wantPlayer2NewElo: 2371,
		},
		{
			name: "2k vs 2.4k player 2 wins",
			args: args{
				player1Elo: 2000,
				player2Elo: 2400,
				winner:     Player2,
			},
			wantPlayer1NewElo: 1997,
			wantPlayer2NewElo: 2403,
		},
		{
			name: "2k vs 2.4k draw",
			args: args{
				player1Elo: 2000,
				player2Elo: 2400,
				winner:     Draw,
			},
			wantPlayer1NewElo: 2013,
			wantPlayer2NewElo: 2387,
		},
		{
			name: "2.4k vs 2k draw",
			args: args{
				player1Elo: 2400,
				player2Elo: 2000,
				winner:     Draw,
			},
			wantPlayer1NewElo: 2387,
			wantPlayer2NewElo: 2013,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPlayer1NewElo, gotPlayer2NewElo := CalculateNewElo(tt.args.player1Elo, tt.args.player2Elo, tt.args.winner)
			if gotPlayer1NewElo != tt.wantPlayer1NewElo {
				t.Errorf("CalculateNewElo() gotPlayer1NewElo = %v, want %v", gotPlayer1NewElo, tt.wantPlayer1NewElo)
			}
			if gotPlayer2NewElo != tt.wantPlayer2NewElo {
				t.Errorf("CalculateNewElo() gotPlayer2NewElo = %v, want %v", gotPlayer2NewElo, tt.wantPlayer2NewElo)
			}
		})
	}
}

// Test_getWinChange runs unit tests for the win chance calculator.
func Test_getWinChance(t *testing.T) {
	type args struct {
		player1Elo int16
		player2Elo int16
	}
	tests := []struct {
		name              string
		args              args
		wantPlayer1Chance float64
		wantPlayer2Chance float64
	}{
		{
			name: "2k vs 2k",
			args: args{
				player1Elo: 2000,
				player2Elo: 2000,
			},
			wantPlayer1Chance: 0.5,
			wantPlayer2Chance: 0.5,
		},
		{
			name: "2k vs 2.8k",
			args: args{
				player1Elo: 2000,
				player2Elo: 2800,
			},
			wantPlayer1Chance: 0.009900990099009901,
			wantPlayer2Chance: 0.9900990099009901,
		},
		{
			name: "2.8k vs 2k",
			args: args{
				player1Elo: 2800,
				player2Elo: 2000,
			},
			wantPlayer1Chance: 0.9900990099009901,
			wantPlayer2Chance: 0.00990099009900991,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPlayer1Chance, gotPlayer2Chance := getWinChance(tt.args.player1Elo, tt.args.player2Elo)
			if gotPlayer1Chance != tt.wantPlayer1Chance {
				t.Errorf("getWinChance() gotPlayer1Chance = %v, want %v", gotPlayer1Chance, tt.wantPlayer1Chance)
			}
			if gotPlayer2Chance != tt.wantPlayer2Chance {
				t.Errorf("getWinChance() gotPlayer2Chance = %v, want %v", gotPlayer2Chance, tt.wantPlayer2Chance)
			}
		})
	}
}

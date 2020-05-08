// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package elo implements an elo calculator, based on https://www.youtube.com/watch?v=AsYfbmp0To0.
package elo

// Player is a typedef for the two different possible players (player 1 or player 2).
type Player uint8

// Winner type enums.
const (
	Draw Player = iota
	Player1
	Player2
)

// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package types defines types and contstants for this application.
package types

// MatchStats is a wrapper for a players match stats, for internal use as a
// dumb container.
type MatchStats struct {
	MMR    int16
	Wins   uint32
	Draws  uint32
	Losses uint32
}

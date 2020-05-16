// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package types defines types and contstants for this application.
package types

import "github.com/6a/blade-ii-api/pkg/elo"

// Structs defined here should also include json serialization hints. They are used to parse request
// bodies that contain data.

// UserCreationRequest describes the request body format for a new user request.
type UserCreationRequest struct {
	Handle   *string `json:"handle"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

// MMRUpdateRequest describes the request body format for an MMR update request.
type MMRUpdateRequest struct {
	Player1ID *uint64     `json:"player1id"`
	Player2ID *uint64     `json:"player2id"`
	Winner    *elo.Player `json:"winner"`
}

// AvatarUpdateRequest describes the request body format for an avatar update request.
type AvatarUpdateRequest struct {
	Avatar    *uint8  `json:"avatar"`
	AuthToken *string `json:"authtoken"`
}

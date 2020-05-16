// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package types defines types and contstants for this application.
package types

// B2ResultCode is a uint16 typedef used for the enumeration of result codes used by this application.
type B2ResultCode uint16

// Code offsets - used to separate codes for different API's.
const (
	OffsetGeneric               = 100
	OffsetCreateAccountHandle   = 200
	OffsetCreateAccountEmail    = 300
	OffsetCreateAccountPassword = 400
	OffsetAuth                  = 500
	OffsetMMR                   = 600
	OffsetGetProfile            = 700
	OffsetUpdateProfile         = 750
	OffsetLeaderboards          = 800
	OffsetGetMatchHistory       = 900
)

// Success indicates that a request was successful.
const Success B2ResultCode = 0

// Generic errors.
const (
	RequestMarshalError B2ResultCode = iota + OffsetGeneric
	ResponseMarshalError
	DatabaseError
	CryptoRandomError
	EmailSendFailure
)

// Create account handle errors.
const (
	HandleMissingOrWrongType B2ResultCode = iota + OffsetCreateAccountHandle
	HandleLength
	HandleSpaceAtStart
	HandleFormat
	HandleAlreadyInUse
	HandleRude
)

// Create account email errors.
const (
	EmailMissingOrWrongType B2ResultCode = iota + OffsetCreateAccountEmail
	EmailFormat
	EmailAlreadyInUse
)

// Create account password errors.
const (
	PasswordMissingOrWrongType B2ResultCode = iota + OffsetCreateAccountPassword
	PasswordComplexityInsufficient
	PasswordFormat
)

// Auth errors.
const (
	AuthHeaderMissing B2ResultCode = iota + OffsetAuth
	AuthHeaderFormat
	AuthInsufficientPermissions
	AuthUsernameOrPasswordIncorrect
	AuthPrivilegeInsufficient
	AuthTokenAuthFailed
	AuthTokenUserNotFound
)

// Update MMR errors.
const (
	Player1IDMissingOrWrongType B2ResultCode = iota + OffsetMMR
	Player2IDMissingOrWrongType
	WinnerMissingOrWrongType
)

// Get Profile errors.
const (
	ProfileGetPublicIDMising B2ResultCode = iota + OffsetGetProfile
	ProfileGetPublicIDNotFound
)

// Get Leaderboards errors.
const (
	LeaderboardsRangeFromMissing B2ResultCode = iota + OffsetLeaderboards
	LeaderboardsRangeFromInvalid
	LeaderboardsRangeCountMissing
	LeaderboardsRangeCountInvalid
)

// Get Match History errors.
const (
	MatchHistoryGetPublicIDMising B2ResultCode = iota + OffsetGetMatchHistory
	MatchHistoryGetPublicIDNotFound
)

// Update profile errors.
const (
	ProfileAvatarUpdateAvatarMissing B2ResultCode = iota + OffsetUpdateProfile
	ProfileAvatarUpdateAuthTokenMissing
	ProfileAvatarUpdateAvatarValueInvalid
)

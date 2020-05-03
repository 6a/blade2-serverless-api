package types

// B2ResultCode is a uint16 typedef used for the enumeration of result codes from the model layer
type B2ResultCode uint16

// code offsets
const (
	OffsetGeneric               = 100
	OffsetCreateAccountHandle   = 200
	OffsetCreateAccountEmail    = 300
	OffsetCreateAccountPassword = 400
	OffsetAuth                  = 500
	OffsetMMR                   = 600
	OffsetGetProfile            = 700
	OffsetLeaderboards          = 800
	OffsetGetMatchHistory       = 900
)

// Success - does this really need a comment, linterさん?
const Success B2ResultCode = iota

// Generic errors
const (
	RequestMarshalError B2ResultCode = iota + OffsetGeneric
	ResponseMarshalError
	DatabaseError
	CryptoRandomError
	EmailSendFailure
)

// Create account handle errors
const (
	HandleMissingOrWrongType B2ResultCode = iota + OffsetCreateAccountHandle
	HandleLength
	HandleSpaceAtStart
	HandleFormat
	HandleAlreadyInUse
	HandleRude
)

// Create account email errors
const (
	EmailMissingOrWrongType B2ResultCode = iota + OffsetCreateAccountEmail
	EmailFormat
	EmailAlreadyInUse
)

// Create account password errors
const (
	PasswordMissingOrWrongType B2ResultCode = iota + OffsetCreateAccountPassword
	PasswordComplexityInsufficient
	PasswordFormat
)

// Auth
const (
	AuthHeaderMissing B2ResultCode = iota + OffsetAuth
	AuthHeaderFormat
	AuthInsufficientPermissions
	AuthUsernameOrPasswordIncorrect
	AuthPrivilegeInsufficient
)

// Update MMR errors
const (
	Player1IDMissingOrWrongType B2ResultCode = iota + OffsetMMR
	Player2IDMissingOrWrongType
	WinnerMissingOrWrongType
)

// Get Profile errors
const (
	ProfileGetPublicIDMising B2ResultCode = iota + OffsetGetProfile
	ProfileGetPublicIDNotFound
)

// Get Leaderboards errors
const (
	LeaderboardsRangeFromMissing B2ResultCode = iota + OffsetLeaderboards
	LeaderboardsRangeFromInvalid
	LeaderboardsRangeCountMissing
	LeaderboardsRangeCountInvalid
)

// Get Match History errors
const (
	MatchHistoryGetPublicIDMising B2ResultCode = iota + OffsetGetMatchHistory
	MatchHistoryGetPublicIDNotFound
)

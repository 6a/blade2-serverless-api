// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package validation is a utility package that contains various validation regex patterns.
package validation

import "regexp"

// Const values for validation.
const (
	UsernameMinLength         = 2
	UsernameMaxLength         = 20
	PasswordMinLengthStandard = 8
	PasswordMinLengthLong     = 15
)

// Regex patterns.
var (
	NoSpaceAtStart         = regexp.MustCompile("^[^\\s]+")
	ValidUsernameRegex     = regexp.MustCompile("^[ー一-龯ぁ-ゞァ-ヶ --/-9Ａ-ｚ０-９！-／;-~：-＠［-｀｛-～、-〜“”‘’´・　]+$")
	ValidEmail             = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	ValidPasswordChars     = regexp.MustCompile("^[ -~]+$")
	NumberAtAnyPosition    = regexp.MustCompile("[0-9]")
	LowerCaseAtAnyPosition = regexp.MustCompile("[a-z]")
)

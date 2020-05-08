// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package database provides an interface through which the application can interact with a database.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/6a/blade-ii-api/internal/settings"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/6a/blade-ii-api/pkg/elo"
	"github.com/6a/blade-ii-game-server/pkg/rid"

	"github.com/alexedwards/argon2id"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/rs/xid"
)

// passwordCheckConstantTimeMin is the minimum amount of time that password check should take - an attemp to make it constant time.
// The argon2id hash check should never take this long, so in theory all auth checks should take this amount of time. This may be
// flawed but I'm not sure... I'm not a netsec expert.
const passwordCheckConstantTimeMin = time.Millisecond * 1500

var (
	// db is a pointer to this packages single instance of a database connection.
	db *sql.DB

	// argonParams are the parameters used for encryption of passwords using
	// the argon2id package.
	argonParams = argon2id.Params{
		Memory:      32 * 1024,
		Iterations:  3,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   30,
	}

	// The following values are simply read from the environment variables, to be used throughout this package
	// in order to interact with the database.
	dbuser          = os.Getenv("db_user")
	dbpass          = os.Getenv("db_pass")
	dburl           = os.Getenv("db_url")
	dbport          = os.Getenv("db_port")
	dbname          = os.Getenv("db_name")
	dbtableUsers    = os.Getenv("db_table_users")
	dbtableProfiles = os.Getenv("db_table_profiles")
	dbtableMatches  = os.Getenv("db_table_matches")
	dbtableTokens   = os.Getenv("db_table_tokens")
)

// Privilege levels for accounts within the database.
const (
	UserPrivilege        uint8 = 0
	GameAdminPrivilege   uint8 = 1
	ServerAdminPrivilege uint8 = 2
)

// Insert a new row into the users table, setting "public_id", "handle", "email", and "salted_hash" with the specified values.
var psCreateAccount = fmt.Sprintf("INSERT INTO `%v`.`%v` (`public_id`, `handle`, `email`, `salted_hash`) VALUES (?, ?, ?, ?);", dbname, dbtableUsers)

// Update a row in the tokens table, setting the value and expiry for the specified token. Contains strings that should be replaced with the token column name
// and token expiry column name - use createAddTokenPS().
var psAddTokenWithReplacers = fmt.Sprintf("UPDATE `%v`.`%v` SET `repl_1` = ?, `repl_2` = DATE_ADD(NOW(), INTERVAL ? HOUR) WHERE `id` = ?;", dbname, dbtableTokens)

// Get the "id" column from the row in the users table with the specified public ID.
var psGetDBIDFromPID = fmt.Sprintf("SELECT `id` FROM `%v`.`%v` WHERE `public_id` = ?;", dbname, dbtableUsers)

// Return a row with a value of either true of false, based on whether a row exists in the users table with the specified handle.
var psCheckName = fmt.Sprintf("SELECT EXISTS(SELECT * FROM `%v`.`%v` WHERE `handle` = ?);", dbname, dbtableUsers)

// Get the "salted_hash", and "banned" column from the row in the users table with the specified handle.
var psCheckAuth = fmt.Sprintf("SELECT `salted_hash`, `banned` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)

// Get the "id", and "public_id" columns from the row in the users table with the specified handle.
var psGetIDs = fmt.Sprintf("SELECT `id`, `public_id` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)

// Get the "privilege" column from the row in the users table with the specified handle.
var psGetPrivilege = fmt.Sprintf("SELECT `privilege` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)

// Insert a new row into the tokens table, setting "id", "email_confirmation", and "email_confirmation_expiry" with the specified values.
var psCreateTokenRowWithEmailToken = fmt.Sprintf("INSERT INTO `%v`.`%v` (`id`, `email_confirmation`, `email_confirmation_expiry`) VALUES (LAST_INSERT_ID(), ?, DATE_ADD(NOW(), INTERVAL ? HOUR));", dbname, dbtableTokens)

// Get the "mmr", "wins", "draws", and "losses" columns from the row in the profiles table with the specified database ID.
var psGetMatchStats = fmt.Sprintf("SELECT `mmr`, `wins`, `draws`, `losses` FROM `%v`.`%v` WHERE `id` = ?;", dbname, dbtableProfiles)

// Update the "mmr", "wins", "draws", and "losses" column for the row in the profiles table with the specified database ID.
var psUpdateMMR = fmt.Sprintf("UPDATE `%v`.`%v` SET `mmr` = ?, `wins` = ?, `draws` = ?, `losses` = ? WHERE `id` = ?;", dbname, dbtableProfiles)

// Get the "avatar", "mmr", "wins", "draws", "losses", "winratio", "ranked_total", and "created" columns from the row in the profiles table with the specified database ID.
var psGetProfile = fmt.Sprintf("SELECT `avatar`, `mmr`, `wins`, `draws`, `losses`, `winratio`, `ranked_total`, `created` FROM `%v`.`%v` WHERE `id` = ?;", dbname, dbtableProfiles)

// Get the "avatar", "mmr", "wins", "draws", "losses", "winratio", "ranked_total" (as "total"), "public_id" (as "pid"), and a generated column "rank" for a range of results specified, ordered using the rank function based on the entire table. ID's 99 or less are excluded due to being reserved for admin accounts.
var psGetLeaderboards = fmt.Sprintf("SELECT `t`.`avatar`, `t`.`mmr`, `t`.`wins`, `t`.`draws`, `t`.`losses`, `t`.`winratio`, `t`.`total`, `t`.`pid`, RANK() OVER (ORDER BY `t`.`mmr` DESC, `t`.`winratio` DESC) AS `rank` FROM (SELECT `avatar`,  `mmr`, `wins`, `draws`, `losses`, `winratio`, `ranked_total` AS `total`, `public_id` AS `pid`, `id`  FROM `%v`.`%v` WHERE `id` >= 100) AS t ORDER BY `rank` LIMIT ? OFFSET ?;", dbname, dbtableProfiles)

// Get the "avatar", "mmr", "wins", "draws", "losses", "winratio", "ranked_total" (as "total"), "public_id" (as "pid"), and a generated column "rank" for the row with the specified public ID, ordered using the rank function based on the entire table. ID's 99 or less are excluded due to being reserved for admin accounts.
var psGetIndividualRank = fmt.Sprintf("SELECT * FROM (SELECT `t`.`avatar`, `t`.`mmr`, `t`.`wins`, `t`.`draws`, `t`.`losses`, `t`.`winratio`, `t`.`total`, `t`.`pid`, RANK() OVER (ORDER BY `t`.`mmr` DESC, `t`.`winratio` DESC) AS `rank` FROM (SELECT `avatar`,  `mmr`, `wins`, `draws`, `losses`, `winratio`, `ranked_total` AS `total`, `public_id` AS `pid`, `id` FROM `%v`.`%v` WHERE `id` >= 100) AS t) AS rt WHERE `pid` = ?;", dbname, dbtableProfiles)

// Get the size of the leaderboards table a single row with a single column. ID's 99 or less are excluded due to being reserved for admin accounts.
var psGetLeaderboardsCount = fmt.Sprintf("SELECT COUNT(*) FROM `%v`.`%v` WHERE `id` >= 100;", dbname, dbtableProfiles)

// Using multiple joins, get the "id" (the match ID), and then "handle" (as "playerNhandle"), and "public_id" (as "playerNpid") for player 1 and 2 respectively, followed by "winnerhandle" and "winnerpid" (from "handle" and "public_id" for the winner), and finally the "end" column, for all of the matches that the specified player took part in, ordered by end time and match ID, in descending order.
var psGetMatchHistory = fmt.Sprintf("SELECT `m`.`id`, `p1`.`handle` as `player1handle`, `p1`.`public_id` as `player1pid`, `p2`.`handle` as `player2handle`, `p2`.`public_id` as `player2pid`, `w`.`handle` as `winnerhandle`, `w`.`public_id` as `winnerpid`, `m`.`end` FROM `%[1]v`.`%[2]v` `m` JOIN `%[1]v`.`%[3]v` `p1` on `p1`.`id` = `m`.`player1` JOIN `%[1]v`.`%[3]v` `p2` on `p2`.`id` = `m`.`player2` JOIN `%[1]v`.`%[3]v` `w` on `w`.`id` = `m`.`winner` WHERE ? IN(`player1`, `player2`) AND `phase` = 2 ORDER BY `end` DESC, `id` DESC;", dbname, dbtableMatches, dbtableUsers)

// Init should be called at the start of the function. It opens a connection to the database
// based on the parameters defined by environment variables, as specified by the EnvironmentVariables struct.
func Init() {

	// Construct the connection string for the database connection.
	connString := fmt.Sprintf("%v:%v@(%v:%v)/%v?tls=skip-verify&parseTime=true", dbuser, dbpass, dburl, dbport, dbname)

	// Attempt to open the connection based on the connection string above. Failure here will
	// cause a panic, as the server cannot function is the database instance is not valid.
	// The resultant database object when successful is stored in the instance variable for this package.
	// Note - an error is declared and used beforehand so that the db can be written to directly, without
	// using a temporary variable.
	var err error
	db, err = sql.Open("mysql", connString)
	if err != nil {
		log.Fatal(err)
	}
}

// CreateUser creates a user account using the specified credentialand and email address. Returns the
// email validation token that was generated upon creation.
func CreateUser(handle string, email string, password string) (emailValidationToken string, err error) {

	// check that the specified user actually exists - early exit on database error.
	exists, err := userExists(handle)
	if err != nil {
		return "", err
	}

	// If the user already exists, return an appropriate error.
	if exists {
		return "", fmt.Errorf("Error 1062: Duplicate entry '%v' for key 'handle_UNIQUE'", handle)
	}

	// As this database interaction has multiple steps, begin a transaction to protect against race conditions.
	transaction, err := db.Begin()

	// Defer rollback of this transaction so that it is cleaned up properly when this function exits.
	defer transaction.Rollback()

	// Exit if there was an error.
	if err != nil {
		return "", err
	}

	// Prepare a statement that will create an account with the specified user details. Exit early on error.
	statement, err := db.Prepare(psCreateAccount)
	if err != nil {
		return "", err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Created a salted password hash of the specified password. Exit early on error.
	saltedhash, err := argon2id.CreateHash(password, &argonParams)
	if err != nil {
		return "", err
	}

	// Create a new UUID, to be used as the public ID for this user.
	publicID := xid.New()

	// Create the user account using the specified, and generated details. Exit early on error.
	_, err = statement.Exec(publicID, handle, email, saltedhash)
	if err != nil {
		return "", err
	}

	// Create a random string (crypto safe) to use as the email confirmation token. Exit early on error.
	emailConfirmationToken, err := rid.RandomString(settings.EmailConfirmationTokenLength)
	if err != nil {
		return "", err
	}

	// Prepare a statement that will add a new row to the tokens table for this user - the id from the last
	// insert is used as the id for the new row. Exit early on error.
	statement, err = db.Prepare(psCreateTokenRowWithEmailToken)
	if err != nil {
		return "", err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Create the new row in the tokens table with the specified settings. Exit early on error.
	_, err = statement.Exec(emailConfirmationToken, settings.EmailConfirmationTokenLifetime)
	if err != nil {
		return "", err
	}

	// Commit the transaction, essentially finalizing all the changes that were just made. Exit early on error.
	err = transaction.Commit()
	if err != nil {
		return "", err
	}

	// Return the email confirmation token, with a nil error.
	return emailConfirmationToken, err
}

// ValidateCredentials checks if the provided handle exists, the user is not banned, and if the hashed password matches
// the stored hashed password.
func ValidateCredentials(handle string, password string) (err error) {

	// Store the start time, so that this function can be made to run in constant time.
	startTime := time.Now()

	// Defer the blocking, makeConstantTime function, so that it fires once this function exits, preventing executing from
	// returning to the caller until the timeout has been reached.
	defer makeConstantTime(startTime, passwordCheckConstantTimeMin)

	// check that the specified user actually exists - early exit on database error.
	exists, err := userExists(handle)
	if err != nil {
		return err
	}

	// If the user does not exist, return an error.
	if !exists {
		return fmt.Errorf("The handle %v does not does not match any existing users", handle)
	}

	// Prepare a statement that will get the salted hash and ban flag for the account with the specified handle.
	statement, err := db.Prepare(psCheckAuth)
	if err != nil {
		return err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the database with the specified handle. Exit early on error.
	var banned bool
	var saltedHash string
	err = statement.QueryRow(handle).Scan(&saltedHash, &banned)
	if err != nil {
		return err
	}

	// If the user is banned, return an error.
	if banned {
		return errors.New("Specified user is banned")
	}

	// Using the argon2id package, compare the specified password against the salted hash retrieved
	// from the database. Exit early on error.
	match, err := argon2id.ComparePasswordAndHash(password, saltedHash)
	if err != nil {
		return err
	}

	// If the password, when hashed with the settings specified by the salted hash, did not match the stored
	// password hash, the match return value will be set to false - so return an error.
	if !match {
		return errors.New("The provided password does not match the stored password for this user")
	}

	// Returning nil indicates that the credentials, and the account, are valid.
	return nil
}

// GetIDs returns the database and public ID for the specified user.
func GetIDs(handle string) (databaseID int, publicID string, err error) {

	// Prepare a statement that will get the database ID and public ID for the specified user.
	statement, err := db.Prepare(psGetIDs)
	if err != nil {
		return -1, "", err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the users table row with the specified handle. Read the found columns into the return
	// variables for this function.
	// Exit early on error.
	err = statement.QueryRow(handle).Scan(&databaseID, &publicID)
	if err != nil {
		return -1, "", err
	}

	// Return the two ID's.
	return databaseID, publicID, nil
}

// SetToken updates the tokens table with the specified token for the specified user.
func SetToken(id int, t types.Token, token string, hoursValid int) (err error) {

	// Prepare a statement that will update the specified token for the specified user. The
	// token will be valid for (hoursValid) hours. Exit early on error.
	statement, err := db.Prepare(createAddTokenPS(t))
	if err != nil {
		return err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the tokens table, setting the new token and expiry for the row with the specified ID.
	// Exit early on error.
	_, err = statement.Exec(token, hoursValid, id)
	if err != nil {
		return err
	}

	// Return nil to indicate that this process was successful.
	return nil
}

// UpdateMatchStats updates the mmr for the two specified clients, as well as w/d/l stats.
func UpdateMatchStats(client1DatabaseID uint64, client1MatchStats types.MatchStats, client2DatabaseID uint64, client2MatchStats types.MatchStats, winner elo.Player) (err error) {

	// As this database interaction has multiple steps, begin a transaction to protect against race conditions.
	transaction, err := db.Begin()

	if err != nil {
		return err
	}

	// Defer rollback of this transaction so that it is cleaned up properly when this function exits.
	defer transaction.Rollback()

	// Update the match stats for player 1.

	// Update wins, losses, or draws depending on the outcome of the match.
	if winner == elo.Draw {
		client1MatchStats.Draws++
	} else if winner == elo.Player1 {
		client1MatchStats.Wins++
	} else {
		client1MatchStats.Losses++
	}

	// Prepare a statement that will update the row in the profiles table for the player 1. Exit early on error.
	statement, err := db.Prepare(psUpdateMMR)
	if err != nil {
		return err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the database, updating the row in the profiles table with the player 1's database ID. Exit early on error.
	_, err = statement.Exec(client1MatchStats.MMR, client1MatchStats.Wins, client1MatchStats.Draws, client1MatchStats.Losses, client1DatabaseID)
	if err != nil {
		return err
	}

	// Update the match stats for player 2.

	// Update wins, losses, or draws depending on the outcome of the match.
	if winner == elo.Draw {
		client2MatchStats.Draws++
	} else if winner == elo.Player2 {
		client2MatchStats.Wins++
	} else {
		client2MatchStats.Losses++
	}

	// Prepare a statement that will update the row in the profiles table for the player 2. Exit early on error.
	statement, err = db.Prepare(psUpdateMMR)
	if err != nil {
		return err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the database, updating the row in the profiles table with the player 1's database ID. Exit early on error.
	_, err = statement.Exec(client2MatchStats.MMR, client2MatchStats.Wins, client2MatchStats.Draws, client2MatchStats.Losses, client2DatabaseID)
	if err != nil {
		return err
	}

	// Commit the transaction, essentially finalizing all the changes that were just made. Exit early on error.
	err = transaction.Commit()
	if err != nil {
		return err
	}

	// Return nil, indicating that the update was successful.
	return nil
}

// GetMatchStats returns the match stats (mmr, w/d/l) for the specified client.
func GetMatchStats(databaseID uint64) (matchStats types.MatchStats, err error) {

	// Prepare a statement that will get the match stats for the specified user. Exit early on error.
	statement, err := db.Prepare(psGetMatchStats)
	if err != nil {
		return matchStats, err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the row in the profiles table for the specified user, and read the returned columns into the return
	// variables of this function. Exit early on error.
	err = statement.QueryRow(databaseID).Scan(&matchStats.MMR, &matchStats.Wins, &matchStats.Draws, &matchStats.Losses)
	if err != nil {
		return matchStats, err
	}

	return matchStats, nil
}

// HasRequiredPrivilege returns true if the specified user has equal, or higher privilege than the specified level.
func HasRequiredPrivilege(handle string, privilegeLevelToCheck uint8) (isPrivileged bool, err error) {

	// Prepare a statement that will get the privilege level for the specified user. Exit early on error.
	statement, err := db.Prepare(psGetPrivilege)
	if err != nil {
		return false, err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the row in the users database for the specified user, and scan the resulting columns value into
	// a temporary value. Exit early on error.
	var actualPrivilege uint8
	err = statement.QueryRow(handle).Scan(&actualPrivilege)
	if err != nil {
		return false, errors.New("The user specified in the auth header does not exist")
	}

	// Determine if the user "is privileged", by checking if the privilege level returned
	// is greater than or equal to the level that was specified.
	isPrivileged = actualPrivilege >= privilegeLevelToCheck

	return isPrivileged, nil
}

// GetDBID returns the databaseID for the specified public ID.
func GetDBID(publicID string) (databaseID uint64, err error) {

	// Prepare a statement that will get the database ID for the specified user. Exit early on error.
	statement, err := db.Prepare(psGetDBIDFromPID)
	if err != nil {
		return databaseID, err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the row in the users database for the specified user (by public ID), reading the returned columns
	// into the retrun variable. Exit early on error.
	err = statement.QueryRow(publicID).Scan(&databaseID)
	if err != nil {
		return databaseID, err
	}

	return databaseID, err
}

// GetProfile returns profile data for the user with the specified databaseID.
func GetProfile(databaseID uint64) (profile types.ProfileResponsePayload, err error) {

	// Prepare a statement that will get the profile data for the specified user. Exit early on error.
	statement, err := db.Prepare(psGetProfile)
	if err != nil {
		return profile, err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the profiles table with the specified database ID. The returned row should contain all the required
	// profile data, which can be scanned into the return profile variable.
	// Note the edge case for winratio - it is possible for this value to be null, so it is scanned into temporary
	// float32 pointer.
	var winRatio *float32 = nil
	err = statement.QueryRow(databaseID).Scan(&profile.Avatar, &profile.MMR, &profile.Wins, &profile.Draws, &profile.Losses, &winRatio, &profile.RankedTotal, &profile.Created)
	if err != nil {
		return profile, err
	}

	// Handle the winratio edge case, setting the win ratio of the return profile variable accordingly.
	if winRatio == nil {
		profile.WinRatio = 0
	} else {
		profile.WinRatio = *winRatio
	}

	return profile, err
}

// GetLeaderboards returns the leaderboards, starting at rank (start) returning (count) results, as well
// as an extra blob containg the details for the user specified.
func GetLeaderboards(publicID string, start uint64, count uint64) (leaderboards types.LeaderboardResponsePayload, err error) {

	// As this database interaction has multiple steps, begin a transaction to protect against race conditions.
	transaction, err := db.Begin()

	// Defer rollback of this transaction so that it is cleaned up properly when this function exits.
	defer transaction.Rollback()

	// Exit if there was an error.
	if err != nil {
		return leaderboards, err
	}

	// Prepare a statement that will get the number of rows in the leaderboard. Exit early on error.
	statement, err := db.Prepare(psGetLeaderboardsCount)
	if err != nil {
		return leaderboards, err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the database - essentially just returns the number of rows in the profiles table. Read
	// the resultant column into a temporay variable. Exit early on error.
	var leaderboardsCount uint64
	err = statement.QueryRow().Scan(&leaderboardsCount)
	if err != nil {
		return leaderboards, err
	}

	// If a public ID was provided, get the row specific to that user as well.
	if publicID != "" {

		// Prepare a statement that will get leaderboards row for a specific user. Exit early on error.
		statement, err = db.Prepare(psGetIndividualRank)
		if err != nil {
			return leaderboards, err
		}

		// Defer closing of the statement so that it is cleaned up properly when this function exits.
		defer statement.Close()

		// Query the profiles table with the specific ID, reading the columns into the return leaderboards
		// variable's user member. Note the edge case for winratio - it is possible for this value to be
		// null, so it is scanned into temporary float32 pointer.
		var winRatio *float32 = nil
		err = statement.QueryRow(publicID).Scan(
			&leaderboards.User.Avatar,
			&leaderboards.User.MMR,
			&leaderboards.User.Wins,
			&leaderboards.User.Draws,
			&leaderboards.User.Losses,
			&winRatio,
			&leaderboards.User.RankedTotal,
			&leaderboards.User.PublicID,
			&leaderboards.User.Rank,
		)

		// Check for errors from the above query.
		if err != nil {

			// Only report db errors, not found = silently return default values.
			if err != sql.ErrNoRows {
				return leaderboards, err
			}
		} else {

			// Handle the winratio edge case, setting the win ratio of the return profile variable accordingly.
			if winRatio == nil {
				leaderboards.User.WinRatio = 0
			} else {
				leaderboards.User.WinRatio = *winRatio
			}

			// Write the determined win ratio into the return variable for the specific user.
			leaderboards.User.WinRatio = *winRatio

			// Also write the leaderboards size to the retrun variable for the specific user.
			leaderboards.User.OutOf = leaderboardsCount
		}
	}

	// Prepare a statement that will get all the leaderboards, with the start poffset and count specified.
	// Exit early on error.
	statement, err = db.Prepare(psGetLeaderboards)
	if err != nil {
		return leaderboards, err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Execute the leaderboards query, which is fairly complex, though is described above where the
	// prepared staments are defined.
	rows, err := statement.Query(count, start)
	if err != nil {
		return leaderboards, err
	}

	// Make an empty slice of leaderboards in the (Leaderboards) member variable of the return
	// leaderboards variable.
	leaderboards.Leaderboards = make([]types.LeaderboardRow, 0)

	// Defer closing of the rows, so that the resource is released properly when the function exits.
	defer rows.Close()

	// Iterate over all of the rows...
	for rows.Next() {

		// For each row, set the outof variable to the size of the leaderboard.
		row := types.LeaderboardRow{
			OutOf: leaderboardsCount,
		}

		// Scan the current row into the current row created above. Note the edge case for winratio
		// - it is possible for this value to be null, so it is scanned into temporary float32 pointer.
		var winRatio *float32 = nil
		err := rows.Scan(
			&row.Avatar,
			&row.MMR,
			&row.Wins,
			&row.Draws,
			&row.Losses,
			&winRatio,
			&row.RankedTotal,
			&row.PublicID,
			&row.Rank,
		)

		// Exit early if the scan errored.
		if err != nil {
			return leaderboards, err
		}

		// Handle the winratio edge case, setting the win ratio of the return profile variable accordingly.
		if winRatio == nil {
			row.WinRatio = 0
		} else {
			row.WinRatio = *winRatio
		}

		// Add the new leaderboards row to the (Leaderboards) array of the return variable.
		leaderboards.Leaderboards = append(leaderboards.Leaderboards, row)
	}

	return leaderboards, nil
}

// GetMatchHistory returns the match history for the specified user.
func GetMatchHistory(databaseID uint64) (history types.MatchHistory, err error) {

	// Prepare a statement that will get the match history for the specified user. Exit early on error.
	statement, err := db.Prepare(psGetMatchHistory)
	if err != nil {
		return history, err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the matches table for the rows with that represent a match that the specified
	// was part of.
	rows, err := statement.Query(databaseID)
	if err != nil {
		return history, err
	}

	// Make an empty slice of match history rows in the return variable.
	history.Rows = make([]types.MatchHistoryRow, 0)

	// Defer closing of the rows, so that the resource is released properly when the function exits.
	defer rows.Close()

	// Iterate over all the rows...
	for rows.Next() {

		// make a new empty match history row.
		row := types.MatchHistoryRow{}

		// Scan all the columns into the new row.
		err := rows.Scan(
			&row.MatchID,
			&row.Player1Handle,
			&row.Player1PublicID,
			&row.Player2Handle,
			&row.Player2PublicID,
			&row.WinnerHandle,
			&row.WinnerPublicID,
			&row.EndTime,
		)

		// Exit early on error.
		if err != nil {
			return history, err
		}

		// Append the new row to the return variables (Rows) member variable.
		history.Rows = append(history.Rows, row)
	}

	return history, nil
}

// createAddTokenPS is a helper function that returns a string where the two
// replacement patterns in the add token prepared statement are switched out,
// based on the type of token that will be written.
func createAddTokenPS(t types.Token) (ps string) {

	// Replace the token type column name.
	ps = strings.Replace(psAddTokenWithReplacers, "repl_1", t.String(), -1)

	// Replace the expiry column name.
	ps = strings.Replace(ps, "repl_2", fmt.Sprintf("%v_expiry", t.String()), -1)

	return ps
}

// userExists retruns true if the user with the specified handle exists.
func userExists(handle string) (exists bool, err error) {

	// Prepare a statement that will check for the existence of the specified user. Exit early on error.
	statement, err := db.Prepare(psCheckName)
	if err != nil {
		return false, err
	}

	// Defer closing of the statement so that it is cleaned up properly when this function exits.
	defer statement.Close()

	// Query the users table for the row with the specified handle. Scan the result into the return variable,
	// and exit early on error.
	err = statement.QueryRow(handle).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return exists, nil
}

// makeConstantTime sleeps until the desired time since the start time has elapsed.
// If the time has already passed, this function immediately returns.
func makeConstantTime(startTime time.Time, desiredTime time.Duration) {
	time.Sleep(desiredTime - time.Since(startTime))
}

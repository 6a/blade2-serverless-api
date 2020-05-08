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

var db *sql.DB

const passwordCheckConstantTimeMin = time.Millisecond * 1500

var argonParams = argon2id.Params{
	Memory:      32 * 1024,
	Iterations:  3,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   30,
}

var dbuser = os.Getenv("db_user")
var dbpass = os.Getenv("db_pass")
var dburl = os.Getenv("db_url")
var dbport = os.Getenv("db_port")
var dbname = os.Getenv("db_name")
var dbtableUsers = os.Getenv("db_table_users")
var dbtableProfiles = os.Getenv("db_table_profiles")
var dbtableMatches = os.Getenv("db_table_matches")
var dbtableTokens = os.Getenv("db_table_tokens")

// Privilege levels
const (
	UserPrivilege        uint8 = 0
	GameAdminPrivilege   uint8 = 1
	ServerAdminPrivilege uint8 = 2
)

var psCreateAccount = fmt.Sprintf("INSERT INTO `%v`.`%v` (`public_id`, `handle`, `email`, `salted_hash`) VALUES (?, ?, ?, ?);", dbname, dbtableUsers)
var psCreateTokenRowWithEmailToken = fmt.Sprintf("INSERT INTO `%v`.`%v` (`id`, `email_confirmation`, `email_confirmation_expiry`) VALUES (LAST_INSERT_ID(), ?, DATE_ADD(NOW(), INTERVAL ? HOUR));", dbname, dbtableTokens)
var psAddTokenWithReplacers = fmt.Sprintf("UPDATE `%v`.`%v` SET `repl_1` = ?, `repl_2` = DATE_ADD(NOW(), INTERVAL ? HOUR) WHERE `id` = ?;", dbname, dbtableTokens)

var psGetDBIDFromPID = fmt.Sprintf("SELECT `id` FROM `%v`.`%v` WHERE `public_id` = ?;", dbname, dbtableUsers)
var psCheckName = fmt.Sprintf("SELECT EXISTS(SELECT * FROM `%v`.`%v` WHERE `handle` = ?);", dbname, dbtableUsers)
var psCheckAuth = fmt.Sprintf("SELECT `salted_hash`, `banned` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)
var psGetIDs = fmt.Sprintf("SELECT `id`, `public_id` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)
var psGetPrivilege = fmt.Sprintf("SELECT `privilege` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)

var psGetMatchStats = fmt.Sprintf("SELECT `mmr`, `wins`, `draws`, `losses` FROM `%v`.`%v` WHERE `id` = ?;", dbname, dbtableProfiles)
var psUpdateMMR = fmt.Sprintf("UPDATE `%v`.`%v` SET `mmr` = ?, `wins` = ?, `draws` = ?, `losses` = ? WHERE `id` = ?;", dbname, dbtableProfiles)
var psGetProfile = fmt.Sprintf("SELECT `avatar`, `mmr`, `wins`, `draws`, `losses`, `winratio`, `ranked_total`, `created` FROM `%v`.`%v` WHERE `id` = ?;", dbname, dbtableProfiles)
var psGetLeaderboards = fmt.Sprintf("SELECT `t`.`avatar`, `t`.`mmr`, `t`.`wins`, `t`.`draws`, `t`.`losses`, `t`.`winratio`, `t`.`total`, `t`.`pid`, RANK() OVER (ORDER BY `t`.`mmr` DESC, `t`.`winratio` DESC) AS `rank` FROM (SELECT `avatar`,  `mmr`, `wins`, `draws`, `losses`, `winratio`, `ranked_total` AS `total`, `public_id` AS `pid`, `id`  FROM `%v`.`%v` WHERE `id` > 100) AS t ORDER BY `rank` LIMIT ? OFFSET ?;", dbname, dbtableProfiles)
var psGetIndividualRank = fmt.Sprintf("SELECT * FROM (SELECT `t`.`avatar`, `t`.`mmr`, `t`.`wins`, `t`.`draws`, `t`.`losses`, `t`.`winratio`, `t`.`total`, `t`.`pid`, RANK() OVER (ORDER BY `t`.`mmr` DESC, `t`.`winratio` DESC) AS `rank` FROM (SELECT `avatar`,  `mmr`, `wins`, `draws`, `losses`, `winratio`, `ranked_total` AS `total`, `public_id` AS `pid`, `id` FROM `%v`.`%v` WHERE `id` > 100) AS t) AS rt WHERE `pid` = ?;", dbname, dbtableProfiles)
var psGetLeaderboardsCount = fmt.Sprintf("SELECT COUNT(*) FROM `%v`.`%v` WHERE `id` > 100;", dbname, dbtableProfiles)

// This one is uber long but it cant be multi-lined due to containing back-ticks
var psGetMatchHistory = fmt.Sprintf("SELECT `m`.`id`, `p1`.`handle` as `player1handle`, `p1`.`public_id` as `player1pid`, `p2`.`handle` as `player2handle`, `p2`.`public_id` as `player2pid`, `w`.`handle` as `winnerhandle`, `w`.`public_id` as `winnerpid`, `m`.`end` FROM `%[1]v`.`%[2]v` `m` JOIN `%[1]v`.`%[3]v` `p1` on `p1`.`id` = `m`.`player1` JOIN `%[1]v`.`%[3]v` `p2` on `p2`.`id` = `m`.`player2` JOIN `%[1]v`.`%[3]v` `w` on `w`.`id` = `m`.`winner` WHERE ? IN(`player1`, `player2`) AND `phase` = 2 ORDER BY `end` DESC, `id` DESC;", dbname, dbtableMatches, dbtableUsers)

var connString = fmt.Sprintf("%v:%v@(%v:%v)/%v?tls=skip-verify&parseTime=true", dbuser, dbpass, dburl, dbport, dbname)

// Init should be called at the start of the function to open a connection to the database
func Init() {
	mysql, err := sql.Open("mysql", connString)
	if err != nil {
		log.Fatal(err)
	}

	db = mysql
}

// CreateUser creates a user account
func CreateUser(handle string, email string, password string) (emailValidationToken string, err error) {
	exists, err := userExists(handle)
	if err != nil {
		return "", err
	}

	if exists {
		return "", fmt.Errorf("Error 1062: Duplicate entry '%v' for key 'handle_UNIQUE'", handle)
	}

	transaction, err := db.Begin()

	defer transaction.Rollback()

	if err != nil {
		return "", err
	}

	statement, err := db.Prepare(psCreateAccount)
	if err != nil {
		return "", err
	}

	defer statement.Close()

	saltedhash, err := argon2id.CreateHash(password, &argonParams)
	if err != nil {
		return "", err
	}

	publicID := xid.New()
	_, err = statement.Exec(publicID, handle, email, saltedhash)
	if err != nil {
		return "", err
	}

	emailConfirmationToken, err := rid.RandomString(settings.EmailConfirmationTokenLength)

	if err != nil {
		return "", err
	}

	statement, err = db.Prepare(psCreateTokenRowWithEmailToken)
	if err != nil {
		return "", err
	}

	defer statement.Close()

	_, err = statement.Exec(emailConfirmationToken, settings.EmailConfirmationTokenLifetime)
	if err != nil {
		return "", err
	}

	err = transaction.Commit()
	if err != nil {
		return "", err
	}

	return emailConfirmationToken, err
}

// ValidateCredentials checks if the provided handle exists, and if the hashed password matches the stored hashed password
func ValidateCredentials(handle string, password string) (err error) {
	startTime := time.Now()
	defer makeConstantTime(startTime, passwordCheckConstantTimeMin)

	exists, err := userExists(handle)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("The handle %v does not does not match any existing users", handle)
	}

	statement, err := db.Prepare(psCheckAuth)
	if err != nil {
		return err
	}

	defer statement.Close()

	var banned bool
	var saltedHash string
	err = statement.QueryRow(handle).Scan(&saltedHash, &banned)
	if err != nil {
		return err
	}

	if banned {
		return errors.New("Specified user is banned")
	}

	match, err := argon2id.ComparePasswordAndHash(password, saltedHash)
	if err != nil {
		return err
	}

	if !match {
		return errors.New("The provided password does not match the stored password for this user")
	}

	return nil
}

// GetIDs returns the database and public ID for the specified user
func GetIDs(handle string) (id int, publicID string, err error) {
	statement, err := db.Prepare(psGetIDs)
	if err != nil {
		return -1, "", err
	}

	defer statement.Close()

	err = statement.QueryRow(handle).Scan(&id, &publicID)
	if err != nil {
		return -1, "", err
	}

	return id, publicID, nil
}

// SetToken updates the tokens table with the specified token for the specified user
func SetToken(id int, t types.Token, token string, hoursValid int) (err error) {
	statement, err := db.Prepare(createAddTokenPS(t))
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(token, hoursValid, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateMMR updates the mmr for the two specified clients, as well as w/d/l stats
func UpdateMMR(client1ID uint64, client1MatchStats types.MatchStats, client2ID uint64, client2MatchStats types.MatchStats, winner elo.Player) (err error) {
	transaction, err := db.Begin()

	if err != nil {
		return err
	}

	defer transaction.Rollback()

	// Player 1
	if winner == elo.Draw {
		client1MatchStats.Draws++
	} else if winner == elo.Player1 {
		client1MatchStats.Wins++
	} else {
		client1MatchStats.Losses++
	}

	statement, err := db.Prepare(psUpdateMMR)
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(client1MatchStats.MMR, client1MatchStats.Wins, client1MatchStats.Draws, client1MatchStats.Losses, client1ID)
	if err != nil {
		return err
	}

	// Player 2
	if winner == elo.Draw {
		client2MatchStats.Draws++
	} else if winner == elo.Player2 {
		client2MatchStats.Wins++
	} else {
		client2MatchStats.Losses++
	}

	statement, err = db.Prepare(psUpdateMMR)
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(client2MatchStats.MMR, client2MatchStats.Wins, client2MatchStats.Draws, client2MatchStats.Losses, client2ID)
	if err != nil {
		return err
	}

	transaction.Commit()
	if err != nil {
		return err
	}

	return err
}

// GetMatchStats returns the match stats (mmr, w/d/l) for the specified client
func GetMatchStats(clientID uint64) (matchStats types.MatchStats, err error) {
	statement, err := db.Prepare(psGetMatchStats)
	if err != nil {
		return matchStats, err
	}

	defer statement.Close()

	err = statement.QueryRow(clientID).Scan(&matchStats.MMR, &matchStats.Wins, &matchStats.Draws, &matchStats.Losses)
	if err != nil {
		return matchStats, err
	}

	return matchStats, nil
}

// HasRequiredPrivilege returns true if the specified user has equal, or higher privilege than the specified level
func HasRequiredPrivilege(handle string, privilegeLevelToCheck uint8) (isPrivileged bool, err error) {
	statement, err := db.Prepare(psGetPrivilege)
	if err != nil {
		return false, err
	}

	defer statement.Close()

	var actualPrivilege uint8
	err = statement.QueryRow(handle).Scan(&actualPrivilege)
	if err != nil {
		return false, errors.New("The user specified in the auth header does not exist")
	}

	isPrivileged = actualPrivilege >= privilegeLevelToCheck

	return isPrivileged, nil
}

// GetDBID returns the DBID for the specified public ID
func GetDBID(publicID string) (DBID uint64, err error) {
	statement, err := db.Prepare(psGetDBIDFromPID)
	if err != nil {
		return DBID, err
	}

	defer statement.Close()

	err = statement.QueryRow(publicID).Scan(&DBID)
	if err != nil {
		return DBID, err
	}

	return DBID, err
}

// GetProfile returns profile data for the user with the specified DBID
func GetProfile(DBID uint64) (profile types.ProfileResponsePayload, err error) {
	statement, err := db.Prepare(psGetProfile)
	if err != nil {
		return profile, err
	}

	defer statement.Close()

	var winRatio *float32 = nil
	err = statement.QueryRow(DBID).Scan(&profile.Avatar, &profile.MMR, &profile.Wins, &profile.Draws, &profile.Losses, &winRatio, &profile.RankedTotal, &profile.Created)
	if err != nil {
		return profile, err
	}

	if winRatio == nil {
		profile.WinRatio = 0
	} else {
		profile.WinRatio = *winRatio
	}

	return profile, err
}

// GetLeaderboards returns the leaderboards, starting at rank <start> returning <count> results, as well
// as an extra blob containg the details for the user specified
func GetLeaderboards(publicID string, start uint64, count uint64) (leaderboards types.LeaderboardResponsePayload, err error) {
	transaction, err := db.Begin()
	if err != nil {
		return leaderboards, err
	}

	defer transaction.Rollback()

	statement, err := db.Prepare(psGetLeaderboardsCount)
	if err != nil {
		return leaderboards, err
	}

	defer statement.Close()

	var leaderboardsCount uint64
	err = statement.QueryRow().Scan(&leaderboardsCount)
	if err != nil {
		return leaderboards, err
	}

	if publicID != "" {
		statement, err = db.Prepare(psGetIndividualRank)
		if err != nil {
			return leaderboards, err
		}

		defer statement.Close()

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

		// Only report db errors, not found = silently return default values
		if err != nil {
			if err != sql.ErrNoRows {
				return leaderboards, err
			}
		} else {
			if winRatio == nil {
				leaderboards.User.WinRatio = 0
			} else {
				leaderboards.User.WinRatio = *winRatio
			}

			leaderboards.User.WinRatio = *winRatio
			leaderboards.User.OutOf = leaderboardsCount
		}
	}

	statement, err = db.Prepare(psGetLeaderboards)
	if err != nil {
		return leaderboards, err
	}

	defer statement.Close()

	rows, err := statement.Query(count, start)
	if err != nil {
		return leaderboards, err
	}

	leaderboards.Leaderboards = make([]types.LeaderboardRow, 0)

	defer rows.Close()

	for rows.Next() {

		row := types.LeaderboardRow{
			OutOf: leaderboardsCount,
		}

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

		if err != nil {
			return leaderboards, err
		}

		if winRatio == nil {
			row.WinRatio = 0
		} else {
			row.WinRatio = *winRatio
		}

		leaderboards.Leaderboards = append(leaderboards.Leaderboards, row)
	}

	return leaderboards, nil
}

// GetMatchHistory returns the match history for the specified user
func GetMatchHistory(DBID uint64) (history types.MatchHistory, err error) {
	statement, err := db.Prepare(psGetMatchHistory)
	if err != nil {
		return history, err
	}

	defer statement.Close()

	rows, err := statement.Query(DBID)
	if err != nil {
		return history, err
	}

	history.Rows = make([]types.MatchHistoryRow, 0)

	defer rows.Close()

	for rows.Next() {

		row := types.MatchHistoryRow{}

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

		if err != nil {
			return history, err
		}

		history.Rows = append(history.Rows, row)
	}

	return history, nil
}

func createAddTokenPS(t types.Token) (ps string) {
	ps = strings.Replace(psAddTokenWithReplacers, "repl_1", t.String(), -1)
	ps = strings.Replace(ps, "repl_2", fmt.Sprintf("%v_expiry", t.String()), -1)

	return ps
}

func userExists(username string) (exists bool, err error) {
	statement, err := db.Prepare(psCheckName)
	if err != nil {
		return false, err
	}

	defer statement.Close()

	err = statement.QueryRow(username).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return exists, nil
}

func makeConstantTime(startTime time.Time, desiredTime time.Duration) {
	time.Sleep(desiredTime - time.Since(startTime))
}

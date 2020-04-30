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
	"github.com/6a/blade-ii-api/pkg/rid"

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

var psCheckName = fmt.Sprintf("SELECT EXISTS(SELECT * FROM `%v`.`%v` WHERE `handle` = ?);", dbname, dbtableUsers)
var psCheckAuth = fmt.Sprintf("SELECT `salted_hash`, `banned` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)
var psGetIDs = fmt.Sprintf("SELECT `id`, `public_id` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)
var psGetPrivilege = fmt.Sprintf("SELECT `privilege` FROM `%v`.`%v` WHERE `handle` = ?;", dbname, dbtableUsers)

var psGetMatchStats = fmt.Sprintf("SELECT `mmr`, `wins`, `draws`, `losses` FROM `%v`.`%v` WHERE `id` = ?;", dbname, dbtableProfiles)
var psUpdateMMR = fmt.Sprintf("UPDATE `%v`.`%v` SET `mmr` = ?, `wins` = ?, `draws` = ?, `losses` = ? WHERE `id` = ?;", dbname, dbtableProfiles)

// var psGetTopN = fmt.Sprintf("SELECT FIND_IN_SET(`wins`, (SELECT GROUP_CONCAT(`wins` ORDER BY `wins` DESC) FROM `%[1]v`.`%[2]v`)) AS `rank`, `name`, `wins`, IFNULL(`winratio`, 0) AS `winratio`, `draws`, `losses`, `played` FROM `%[1]v`.`%[2]v` ORDER BY `rank` = 0, `rank`, `winratio` DESC LIMIT ?;", dbname, dbtable)
// var psGetUser = fmt.Sprintf("SELECT FIND_IN_SET(`wins`, (SELECT GROUP_CONCAT(`wins` ORDER BY `wins` DESC) FROM `%[1]v`.`%[2]v`)) AS `rank`, `wins`, IFNULL(`winratio`, 0) AS `winratio`, `draws`, `losses`, `played` FROM `%[1]v`.`%[2]v` WHERE `name` = ?;", dbname, dbtable)
// var psUpdateUser = fmt.Sprintf("UPDATE `%v`.`%v` SET `wins` = (`wins` + ?), `draws` = (`draws` + ?), `losses` = (`losses` + ?) WHERE `name` = ?;", dbname, dbtable)
// var psCount = fmt.Sprintf("SELECT COUNT(*) FROM `%v`.`%v`;", dbname, dbtable)
var connString = fmt.Sprintf("%v:%v@(%v:%v)/%v?tls=skip-verify", dbuser, dbpass, dburl, dbport, dbname)

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

	if err != nil {
		return "", err
	}

	statement, err := db.Prepare(psCreateAccount)
	if err != nil {
		transaction.Rollback()
		return "", err
	}

	defer statement.Close()

	saltedhash, err := argon2id.CreateHash(password, &argonParams)
	if err != nil {
		transaction.Rollback()
		return "", err
	}

	publicID := xid.New()
	_, err = statement.Exec(publicID, handle, email, saltedhash)
	if err != nil {
		transaction.Rollback()
		return "", err
	}

	emailConfirmationToken, err := rid.RandomString(settings.EmailConfirmationTokenLength)

	if err != nil {
		transaction.Rollback()
		return "", err
	}

	statement, err = db.Prepare(psCreateTokenRowWithEmailToken)
	if err != nil {
		transaction.Rollback()
		return "", err
	}

	_, err = statement.Exec(emailConfirmationToken, settings.EmailConfirmationTokenLifetime)
	if err != nil {
		transaction.Rollback()
		return "", err
	}

	transaction.Commit()

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
		transaction.Rollback()
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(client1MatchStats.MMR, client1MatchStats.Wins, client1MatchStats.Draws, client1MatchStats.Losses, client1ID)
	if err != nil {
		transaction.Rollback()
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
		transaction.Rollback()
		return err
	}

	_, err = statement.Exec(client2MatchStats.MMR, client2MatchStats.Wins, client2MatchStats.Draws, client2MatchStats.Losses, client2ID)
	if err != nil {
		transaction.Rollback()
		return err
	}

	transaction.Commit()

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

// // GetLeaderboard returns the leaderboard info, asligned with the specified user and capped to n results
// func GetLeaderboard(username string, maxResults int) (leaderboard types.Leaderboard, err error) {
// 	statement, err := db.Prepare(psGetUser)
// 	if err != nil {
// 		return leaderboard, err
// 	}

// 	defer statement.Close()

// 	var (
// 		name   string
// 		rank   int
// 		wins   int
// 		ratio  float32
// 		draws  int
// 		losses int
// 		played int
// 	)

// 	err = statement.QueryRow(username).Scan(&rank, &wins, &ratio, &draws, &losses, &played)

// 	if err != nil {
// 		return leaderboard, err
// 	}

// 	statement, err = db.Prepare(psCount)
// 	if err != nil {
// 		return leaderboard, err
// 	}

// 	defer statement.Close()

// 	var outof int
// 	err = statement.QueryRow().Scan(&outof)
// 	if err != nil {
// 		return leaderboard, err
// 	}

// 	leaderboard.User.Fill(username, rank, outof, wins, ratio, draws, losses, played)

// 	statement, err = db.Prepare(psGetTopN)
// 	if err != nil {
// 		return leaderboard, err
// 	}

// 	defer statement.Close()
// 	rows, err := statement.Query(maxResults)
// 	if err != nil {
// 		return leaderboard, err
// 	}

// 	leaderboard.Leaderboard = make([]types.LeaderboardRow, 0)

// 	defer rows.Close()
// 	for rows.Next() {
// 		err := rows.Scan(&rank, &name, &wins, &ratio, &draws, &losses, &played)
// 		if err != nil {
// 			return leaderboard, err
// 		}

// 		var row = types.LeaderboardRow{}
// 		row.Fill(name, rank, outof, wins, ratio, draws, losses, played)
// 		leaderboard.Leaderboard = append(leaderboard.Leaderboard, row)
// 	}

// 	return leaderboard, err
// }

// // Update updates the user specified in the input data with the deltas also specified in the input data
// func Update(indata types.UserUpdateRequest) (err error) {
// 	statement, err := db.Prepare(psUpdateUser)
// 	if err != nil {
// 		return err
// 	}

// 	defer statement.Close()

// 	_, err = statement.Exec(indata.Wins, indata.Draws, indata.Losses, indata.Name)
// 	if err != nil {
// 		return err
// 	}

// 	return err
// }

// func validateCredentials(username string, password string) (valid bool, err error) {
// 	statement, err := db.Prepare(psCheckAuth)
// 	if err != nil {
// 		return false, err
// 	}

// 	defer statement.Close()

// 	var banned bool
// 	var saltyhash string
// 	err = statement.QueryRow(username).Scan(&saltyhash, &banned)
// 	if err != nil {
// 		return false, err
// 	}

// 	if banned {
// 		return false, errors.New("Your account has been banned")
// 	}

// 	valid, err = argon2id.ComparePasswordAndHash(password, saltyhash)

// 	return valid, nil
// }

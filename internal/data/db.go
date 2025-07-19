package data

import (
	"database/sql"
	"errors"
	"os"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type User struct {
	Username        string
	PasswordHash    string
	KISAccountEnc   string // encrypted
	KISRealKeyEnc   string // encrypted
	KISRealSecretEnc string // encrypted
	KISPaperKeyEnc   string // encrypted
	KISPaperSecretEnc string // encrypted
}

func CreateUserTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password_hash TEXT NOT NULL,
			kis_account_enc TEXT,
			kis_real_key_enc TEXT,
			kis_real_secret_enc TEXT,
			kis_paper_key_enc TEXT,
			kis_paper_secret_enc TEXT
		);
	`)
	return err
}

func RegisterUser(db *sql.DB, username, passwordHash string) error {
	_, err := db.Exec(`INSERT INTO users (username, password_hash) VALUES ($1, $2)`, username, passwordHash)
	if err != nil {
		// Check for PostgreSQL-specific errors
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return errors.New("username already exists")
			case "23514": // check_violation
				return errors.New("invalid data provided")
			default:
				return errors.New("database error: " + pqErr.Message)
			}
		}
		return errors.New("failed to register user: " + err.Error())
	}
	return nil
}

// Update KIS account info for a user
func UpdateKISAccountInfo(db *sql.DB, username, kisAccountEnc string) error {
	result, err := db.Exec(`UPDATE users SET kis_account_enc = $1 WHERE username = $2`, kisAccountEnc, username)
	if err != nil {
		return errors.New("failed to update KIS account info: " + err.Error())
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check update result: " + err.Error())
	}
	
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	
	return nil
}

func UpdateKISAPIKeys(db *sql.DB, username, realKeyEnc, realSecretEnc, paperKeyEnc, paperSecretEnc string) error {
	result, err := db.Exec(`UPDATE users SET kis_real_key_enc = $1, kis_real_secret_enc = $2, kis_paper_key_enc = $3, kis_paper_secret_enc = $4 WHERE username = $5`, realKeyEnc, realSecretEnc, paperKeyEnc, paperSecretEnc, username)
	if err != nil {
		return errors.New("failed to update KIS API keys: " + err.Error())
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check update result: " + err.Error())
	}
	
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	
	return nil
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	row := db.QueryRow(`SELECT username, password_hash FROM users WHERE username = $1`, username)
	var user User
	if err := row.Scan(&user.Username, &user.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.New("database error: " + err.Error())
	}
	return &user, nil
}

func OpenDB() (*sql.DB, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := CreateUserTable(db); err != nil {
		return nil, err
	}
	return db, nil
} 
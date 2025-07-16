package data

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/lib/pq"
)

type User struct {
	Username     string
	PasswordHash string
}

func CreateUserTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password_hash TEXT NOT NULL
		);
	`)
	return err
}

func RegisterUser(db *sql.DB, username, passwordHash string) error {
	_, err := db.Exec(`INSERT INTO users (username, password_hash) VALUES ($1, $2)`, username, passwordHash)
	if err != nil {
		return err
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
		return nil, err
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
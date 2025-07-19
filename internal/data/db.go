package data

import (
	"database/sql"
	"errors"
	"os"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func RegisterUser(db *sql.DB, username, passwordHash string) error {
	_, err := db.Exec(`INSERT INTO users (username, password_hash) VALUES ($1, $2)`, username, passwordHash)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return errors.New("username already exists")
			case "23514":
				return errors.New("invalid data provided")
			default:
				return errors.New("database error: " + pqErr.Message)
			}
		}
		return errors.New("failed to register user: " + err.Error())
	}
	return nil
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	row := db.QueryRow(`SELECT id, username, password_hash, created_at FROM users WHERE username = $1`, username)
	var user User
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.New("database error: " + err.Error())
	}
	return &user, nil
}

// Add GetUserByID(db, id) function to fetch a user by their ID
func GetUserByID(db *sql.DB, id int64) (*User, error) {
	row := db.QueryRow(`SELECT id, username, password_hash, created_at FROM users WHERE id = $1`, id)
	var user User
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// MigrateDB creates or updates all tables as per the spec
func MigrateDB(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS user_accounts (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		account_id VARCHAR(20) NOT NULL,
		enc_cano BYTEA NOT NULL,
		enc_app_key BYTEA NOT NULL,
		enc_app_secret BYTEA NOT NULL,
		is_mock BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT NOW(),
		UNIQUE (user_id, account_id)
	);

	CREATE TABLE IF NOT EXISTS kis_access_tokens (
		user_account_id BIGINT PRIMARY KEY REFERENCES user_accounts(id) ON DELETE CASCADE,
		token VARCHAR(512) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		refreshed_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS orders (
		id BIGSERIAL PRIMARY KEY,
		user_account_id BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
		symbol VARCHAR(12) NOT NULL,
		side VARCHAR(4) CHECK (side IN ('BUY','SELL')),
		qty NUMERIC(18,2) NOT NULL,
		order_type VARCHAR(6) CHECK (order_type IN ('MARKET','LIMIT')),
		limit_price NUMERIC(18,2),
		status VARCHAR(12) DEFAULT 'PENDING',
		kis_order_id VARCHAR(64),
		created_at TIMESTAMP DEFAULT NOW()
	);
	`)
	return err
}

// Update OpenDB to call MigrateDB
func OpenDB() (*sql.DB, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := MigrateDB(db); err != nil {
		return nil, err
	}
	return db, nil
}

// User Accounts
func CreateUserAccount(db *sql.DB, ua *UserAccount) error {
	query := `INSERT INTO user_accounts (user_id, account_id, enc_cano, enc_app_key, enc_app_secret, is_mock) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	return db.QueryRow(query, ua.UserID, ua.AccountID, ua.EncCANO, ua.EncAppKey, ua.EncAppSecret, ua.IsMock).Scan(&ua.ID, &ua.CreatedAt)
}

func GetUserAccountsByUserID(db *sql.DB, userID int64) ([]UserAccount, error) {
	rows, err := db.Query(`SELECT id, user_id, account_id, enc_cano, enc_app_key, enc_app_secret, is_mock, created_at FROM user_accounts WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []UserAccount
	for rows.Next() {
		var ua UserAccount
		if err := rows.Scan(&ua.ID, &ua.UserID, &ua.AccountID, &ua.EncCANO, &ua.EncAppKey, &ua.EncAppSecret, &ua.IsMock, &ua.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, ua)
	}
	return accounts, nil
}

func DeleteUserAccount(db *sql.DB, userID, accountID int64) error {
	_, err := db.Exec(`DELETE FROM user_accounts WHERE user_id = $1 AND id = $2`, userID, accountID)
	return err
}

// KIS Access Tokens
func UpsertKISToken(db *sql.DB, t *KISAccessToken) error {
	_, err := db.Exec(`INSERT INTO kis_access_tokens (user_account_id, token, expires_at, refreshed_at) VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_account_id) DO UPDATE SET token = $2, expires_at = $3, refreshed_at = $4`, t.UserAccountID, t.Token, t.ExpiresAt, t.RefreshedAt)
	return err
}

func GetKISToken(db *sql.DB, userAccountID int64) (*KISAccessToken, error) {
	row := db.QueryRow(`SELECT user_account_id, token, expires_at, refreshed_at FROM kis_access_tokens WHERE user_account_id = $1`, userAccountID)
	var t KISAccessToken
	if err := row.Scan(&t.UserAccountID, &t.Token, &t.ExpiresAt, &t.RefreshedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// Orders
func CreateOrder(db *sql.DB, o *Order) error {
	query := `INSERT INTO orders (user_account_id, symbol, side, qty, order_type, limit_price, status, kis_order_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`
	return db.QueryRow(query, o.UserAccountID, o.Symbol, o.Side, o.Qty, o.OrderType, o.LimitPrice, o.Status, o.KISOrderID).Scan(&o.ID, &o.CreatedAt)
}

func GetOrderByID(db *sql.DB, userID, orderID int64) (*Order, error) {
	row := db.QueryRow(`SELECT o.id, o.user_account_id, o.symbol, o.side, o.qty, o.order_type, o.limit_price, o.status, o.kis_order_id, o.created_at FROM orders o JOIN user_accounts ua ON o.user_account_id = ua.id WHERE o.id = $1 AND ua.user_id = $2`, orderID, userID)
	var o Order
	if err := row.Scan(&o.ID, &o.UserAccountID, &o.Symbol, &o.Side, &o.Qty, &o.OrderType, &o.LimitPrice, &o.Status, &o.KISOrderID, &o.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func ListOrdersByAccountID(db *sql.DB, userID, userAccountID int64) ([]Order, error) {
	rows, err := db.Query(`SELECT o.id, o.user_account_id, o.symbol, o.side, o.qty, o.order_type, o.limit_price, o.status, o.kis_order_id, o.created_at FROM orders o JOIN user_accounts ua ON o.user_account_id = ua.id WHERE ua.user_id = $1 AND o.user_account_id = $2`, userID, userAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserAccountID, &o.Symbol, &o.Side, &o.Qty, &o.OrderType, &o.LimitPrice, &o.Status, &o.KISOrderID, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
} 
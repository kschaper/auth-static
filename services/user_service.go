package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// PasswordMinLen is the minimum password length.
const PasswordMinLen = 8

// Error represents an error returned on expected errors.
type Error string

// Error returns the error message
func (e Error) Error() string {
	return string(e)
}

const (
	// ErrEmailRequired is returned when the email is empty
	ErrEmailRequired = Error("email required")
	// ErrPasswordTooShort is returned when the password is < PasswordMinLen.
	ErrPasswordTooShort = Error("password too short")
	// ErrPasswordNotConfirmed is return when password and confirmation don't match.
	ErrPasswordNotConfirmed = Error("password and doesn't match confirmation")
	// ErrUnknownCode is returned when the given code is not in db.
	ErrUnknownCode = Error("code unknown")
)

// UserService manages users.
type UserService struct {
	DB *sql.DB
}

// Create creates a new user with the given email and a generated code which is then returned.
// The code can be used for the signup URL: `https://example.com/signup/<code>`.
// If a user with the given email already exists then it will be updated
// with a new code, an empty password hash, and an updated updated_at.
// All other fields won't get updated.
func (service *UserService) Create(email string) (string, error) {
	// validate email length
	if len(email) == 0 {
		return "", ErrEmailRequired
	}

	// generate new code
	code, err := generateCode()
	if err != nil {
		return "", err
	}

	// create new user
	sql := "INSERT INTO users (id, email, code, created_at) VALUES (?, ?, ?, DATETIME('now')) " +
		"ON CONFLICT(email) DO UPDATE SET code = ?, hash = '', updated_at = DATETIME('now')"
	stmt, err := service.DB.Prepare(sql)
	if err != nil {
		return "", err
	}
	if _, err := stmt.Exec(uuid.NewV4(), email, code, code); err != nil {
		return "", err
	}
	return code, nil
}

// generateCode generates a code.
func generateCode() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetIDByCode returns the user ID for the given code.
func (service *UserService) GetIDByCode(code string) (uuid.UUID, error) {
	stmt, err := service.DB.Prepare("SELECT id FROM users WHERE code = ?")
	if err != nil {
		return uuid.Nil, err
	}

	var id string
	err = stmt.QueryRow(code).Scan(&id)
	if err == sql.ErrNoRows {
		return uuid.Nil, ErrUnknownCode
	}
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.FromString(id)
}

// GetIDByEmail returns the user ID for the given email.
func (service *UserService) GetIDByEmail(email string) (uuid.UUID, error) {
	stmt, err := service.DB.Prepare("SELECT id FROM users WHERE email = ?")
	if err != nil {
		return uuid.Nil, err
	}

	var id string
	err = stmt.QueryRow(email).Scan(&id)
	if err == sql.ErrNoRows {
		return uuid.Nil, ErrUnknownCode
	}
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.FromString(id)
}

// UpdatePassword sets the hash and deletes the code.
func (service *UserService) UpdatePassword(id uuid.UUID, password, confirmation string) error {
	password = strings.TrimSpace(password)
	confirmation = strings.TrimSpace(confirmation)

	// validate minimum password length
	if len(password) < PasswordMinLen {
		return ErrPasswordTooShort
	}

	// validate password confirmation
	if password != confirmation {
		return ErrPasswordNotConfirmed
	}

	// generate password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// update user
	stmt, err := service.DB.Prepare("UPDATE users SET hash = ?, code = '', updated_at = DATETIME('now') WHERE id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(string(hash), id)
	if err != nil {
		return err
	}
	return nil
}

// Authenticate checks if there is a user for the given email and password.
func (service *UserService) Authenticate(email, password string) (bool, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	// get the hashed password
	stmt, err := service.DB.Prepare("SELECT hash FROM users WHERE email = ?")
	if err != nil {
		return false, err
	}

	var hash string
	err = stmt.QueryRow(email).Scan(&hash)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// compare hash and password
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	// password wrong
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}

	// some other error
	if err != nil {
		return false, err
	}
	return true, nil
}

// Exists checks if the user with the given ID exists.
func (service *UserService) Exists(id uuid.UUID) (bool, error) {
	var count int
	err := service.DB.QueryRow("SELECT COUNT(id) FROM users WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

package services_test

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"github.com/kschaper/auth-static/services"
)

func db(t *testing.T) *sql.DB {
	client := &services.DatabaseClient{DSN: ":memory:"}
	db, err := client.Open()
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestUserService_Create(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"valid": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = "me@example.com"
			)

			// create user
			code, err := userService.Create(email)
			if err != nil {
				t.Fatalf("expected no error but got %q", err)
			}

			// ensure returned code is not empty
			if code == "" {
				t.Fatal("expected code not to be empty")
			}

			// get user values from db
			var (
				storedID        string
				storedEmail     string
				storedCreatedAt string
			)

			row := db.QueryRow("SELECT id, email, created_at FROM users WHERE code = $1", code)
			if err := row.Scan(&storedID, &storedEmail, &storedCreatedAt); err != nil {
				t.Fatal(err)
			}

			// ensure id is set
			id, err := uuid.FromString(storedID)
			if err != nil {
				t.Fatal(err)
			}
			if id == uuid.Nil {
				t.Fatal("expected id not to be nil")
			}

			// ensure email is correct
			if storedEmail != email {
				t.Fatalf("expected user to have email %q but has %q", email, storedEmail)
			}

			// ensure created_at is set
			parsedCreatedAt, err := time.Parse("2006-01-02 15:04:05", storedCreatedAt)
			if err != nil {
				t.Fatal(err)
			}
			if parsedCreatedAt.IsZero() {
				t.Fatal("expected created_at not to be zero")
			}
		},
		"empty email": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = ""
			)

			// create user
			if _, err := userService.Create(email); err != services.ErrEmailRequired {
				t.Fatalf("expected error %q but got %q\n", services.ErrEmailRequired, err)
			}

		},
		"duplicate email": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = "me@example.com"
			)

			// create user
			if _, err := userService.Create(email); err != nil {
				t.Fatalf("expected no error but got %q", err)
			}

			// get user
			var (
				firstID        string
				firstCreatedAt string
				firstUpdatedAt sql.NullString
			)

			row := db.QueryRow("SELECT id, created_at, updated_at FROM users WHERE email = $1", email)
			if err := row.Scan(&firstID, &firstCreatedAt, &firstUpdatedAt); err != nil {
				t.Fatal(err)
			}

			// create user again
			code, err := userService.Create(email)
			if err != nil {
				t.Fatalf("expected no error but got %q", err)
			}

			// ensure only one user with this email exists
			var num int
			row = db.QueryRow("SELECT COUNT(id) FROM users WHERE email = $1", email)
			if err := row.Scan(&num); err != nil {
				t.Fatal(err)
			}
			if num != 1 {
				t.Fatalf("expected to get exactly 1 record for this email but got %d\n", num)
			}

			// get user again
			var (
				secondID        string
				secondCode      string
				secondHash      string
				secondCreatedAt string
				secondUpdatedAt sql.NullString
			)

			row = db.QueryRow("SELECT id, code, hash, created_at, updated_at FROM users WHERE email = $1", email)
			if err := row.Scan(&secondID, &secondCode, &secondHash, &secondCreatedAt, &secondUpdatedAt); err != nil {
				t.Fatal(err)
			}

			// ensure id is the same
			if secondID != firstID {
				t.Fatal("expected id to be the same but wasn't")
			}

			// ensure code has been updated
			if secondCode != code {
				t.Fatal("expected code to be updated but wasn't")
			}

			// ensure password hash has been deleted
			if secondHash != "" {
				t.Fatal("expected hash to be empty but wasn't")
			}

			// ensure created_at has not beeen changed
			if secondCreatedAt != firstCreatedAt {
				t.Fatal("expected created_at to be the same but wasn't")
			}

			// ensure updated_at has been changed
			if secondUpdatedAt == firstUpdatedAt {
				t.Fatal("expected updated_at not to be the same but was")
			}
		},
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

func TestUserService_GetIDByCode(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"success": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = "me@example.com"
			)

			// create user
			code, err := userService.Create(email)
			if err != nil {
				t.Fatalf("expected no error but got %q", err)
			}

			// get id by code
			id, err := userService.GetIDByCode(code)
			if err != nil {
				t.Fatal(err)
			}

			// ensure id is correct
			var storedEmail string
			row := db.QueryRow("SELECT email FROM users WHERE id = $1", id)
			if err := row.Scan(&storedEmail); err != nil {
				t.Fatal(err)
			}

			if storedEmail != email {
				t.Fatalf("expected email %q but got %q\n", email, storedEmail)
			}
		},
		"unknown id": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
			)

			// get id by code
			_, err := userService.GetIDByCode("e80ef0a04db3597e09fee4e958ca12b1")
			if err != services.ErrUnknownCode {
				t.Fatalf("expected error %q but got %q\n", services.ErrUnknownCode, err)
			}
		},
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

func TestUserService_GetIDByEmail(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"success": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = "me@example.com"
			)

			// create user
			if _, err := userService.Create(email); err != nil {
				t.Fatal(err)
			}

			// get id by email
			id, err := userService.GetIDByEmail(email)
			if err != nil {
				t.Fatalf("expected no error but got %q", err)
			}

			// ensure id is correct
			var storedEmail string
			row := db.QueryRow("SELECT email FROM users WHERE id = $1", id)
			if err := row.Scan(&storedEmail); err != nil {
				t.Fatal(err)
			}

			if storedEmail != email {
				t.Fatalf("expected email %q but got %q\n", email, storedEmail)
			}
		},
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

func TestUserService_UpdatePassword(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"success": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = "me@example.com"
				password    = strings.Repeat("x", services.PasswordMinLen)
			)

			// create user
			code, err := userService.Create(email)
			if err != nil {
				t.Fatal(err)
			}

			// get user id
			id, err := userService.GetIDByCode(code)
			if err != nil {
				t.Fatal(err)
			}

			// update the password
			if err := userService.UpdatePassword(id, password, password); err != nil {
				t.Fatalf("expected no error but got %q\n", err)
			}

			// ensure hash and updated_at have been set, and code has been deleted
			var (
				storedHash      sql.NullString
				storedCode      string
				storedUpdatedAt string
			)

			row := db.QueryRow("SELECT hash, code, updated_at FROM users WHERE id = $1", id)
			if err := row.Scan(&storedHash, &storedCode, &storedUpdatedAt); err != nil {
				t.Fatal(err)
			}

			if !storedHash.Valid {
				t.Fatal("expected hash not to be nil")
			}

			if storedCode != "" {
				t.Fatal("expected code to be empty but wasn't")
			}

			parsedUpdatedAt, err := time.Parse("2006-01-02 15:04:05", storedUpdatedAt)
			if err != nil {
				t.Fatal(err)
			}
			if parsedUpdatedAt.IsZero() {
				t.Fatal("expected updated_at not to be zero")
			}
		},
		"password too short": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = "me@example.com"
				password    = strings.Repeat("x", services.PasswordMinLen-1)
			)

			// create user
			code, err := userService.Create(email)
			if err != nil {
				t.Fatal(err)
			}

			// get user id
			id, err := userService.GetIDByCode(code)
			if err != nil {
				t.Fatal(err)
			}

			// update the password
			err = userService.UpdatePassword(id, password, password)
			if err == nil || err.Error() != services.ErrPasswordTooShort.Error() {
				t.Fatalf("expected to get error %q but got %q\n", services.ErrPasswordTooShort, err)
			}
		},
		"password confirmation mismatch": func(t *testing.T) {
			var (
				db           = db(t)
				userService  = &services.UserService{DB: db}
				email        = "me@example.com"
				password     = strings.Repeat("x", services.PasswordMinLen)
				confirmation = strings.Repeat("y", services.PasswordMinLen)
			)

			// create user
			code, err := userService.Create(email)
			if err != nil {
				t.Fatal(err)
			}

			// get user id
			id, err := userService.GetIDByCode(code)
			if err != nil {
				t.Fatal(err)
			}

			// update the password
			err = userService.UpdatePassword(id, password, confirmation)
			if err == nil || err.Error() != services.ErrPasswordNotConfirmed.Error() {
				t.Fatalf("expected to get error %q but got %q\n", services.ErrPasswordNotConfirmed, err)
			}
		},
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

func TestUserService_Authenticate(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"success": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = "me@example.com"
				password    = strings.Repeat("x", services.PasswordMinLen)
			)

			// create user
			code, err := userService.Create(email)
			if err != nil {
				t.Fatal(err)
			}

			// get user id
			id, err := userService.GetIDByCode(code)
			if err != nil {
				t.Fatal(err)
			}

			// update the password
			if err = userService.UpdatePassword(id, password, password); err != nil {
				t.Fatal(err)
			}

			// authenticate
			authenticated, err := userService.Authenticate(email, password)
			if err != nil {
				t.Fatalf("expected no error but got %q\n", err)
			}
			if !authenticated {
				t.Fatal("expected user to be authenticated but wasn't")
			}
		},
		"unknown email": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
			)

			// authenticate
			authenticated, err := userService.Authenticate("unknown@example.com", strings.Repeat("x", services.PasswordMinLen))
			if err != nil {
				t.Fatal(err)
			}
			if authenticated {
				t.Fatal("expected user not to be authenticated but was")
			}
		},
		"wrong password": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = "me@example.com"
				password    = strings.Repeat("x", services.PasswordMinLen)
			)

			// create user
			code, err := userService.Create(email)
			if err != nil {
				t.Fatal(err)
			}

			// get user id
			id, err := userService.GetIDByCode(code)
			if err != nil {
				t.Fatal(err)
			}

			// update the password
			if err = userService.UpdatePassword(id, password, password); err != nil {
				t.Fatal(err)
			}

			// authenticate
			authenticated, err := userService.Authenticate(email, strings.Repeat("y", services.PasswordMinLen))
			if err != nil {
				t.Fatal(err)
			}
			if authenticated {
				t.Fatal("expected user not to be authenticated but was")
			}
		},
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

func TestUserService_Exists(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"user does exist": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
			)

			// create user
			code, err := userService.Create("me@example.com")
			if err != nil {
				t.Fatal(err)
			}

			// get user id
			id, err := userService.GetIDByCode(code)
			if err != nil {
				t.Fatal(err)
			}

			// check if user exits
			exists, err := userService.Exists(id)
			if err != nil {
				t.Fatalf("expected no error but got %q\n", err)
			}
			if !exists {
				t.Fatal("expected user to exist but it doesn't")
			}
		},
		"user does not exist": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
			)

			// check if non-existent user exits
			exists, err := userService.Exists(uuid.NewV4())
			if err != nil {
				t.Fatalf("expected no error but got %q\n", err)
			}
			if exists {
				t.Fatal("expected the user not to exist but it does")
			}
		},
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

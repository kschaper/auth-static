package handlers_test

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"

	"gitlab.com/kschaper/auth-static/handlers"
	"gitlab.com/kschaper/auth-static/services"
)

func db(t *testing.T) *sql.DB {
	client := &services.DatabaseClient{DSN: ":memory:"}
	db, err := client.Open()
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSignupFormHandler(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"success": func(t *testing.T) {
			code := "73d3e3502ab73f40d4943fdcc16d05dd"

			// server
			store := sessions.NewCookieStore([]byte("abc"))
			mux := http.NewServeMux()
			mux.HandleFunc("/signup/", handlers.SignupFormHandler(store))
			ts := httptest.NewServer(mux)
			defer ts.Close()

			// request
			req, err := http.Get(ts.URL + "/signup/" + code)
			if err != nil {
				t.Fatal(err)
			}
			defer req.Body.Close()

			// ensure status code 200
			if req.StatusCode != http.StatusOK {
				t.Fatalf("expected status code %d but got %d\n", http.StatusOK, req.StatusCode)
			}

			// ensure code is in HTML
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			html := string(body)
			expected := fmt.Sprintf(`action="/signup/%s"`, code)
			match, err := regexp.MatchString(expected, html)
			if err != nil {
				t.Fatal(err)
			}
			if !match {
				t.Fatalf("expected html to contain\n%s\nbut didn't:\n%s\n", expected, html)
			}
		},
		// TODO: test rendering of error messages
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

func TestSignupHandler(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"success": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
				email       = "webmaster@example.com"
				password    = strings.Repeat("k", services.PasswordMinLen)
			)

			// create user
			code, err := userService.Create(email)
			if err != nil {
				t.Fatalf("expected no error but gut %q", err)
			}

			// server
			store := sessions.NewCookieStore([]byte("abc"))
			mux := http.NewServeMux()
			mux.HandleFunc("/signup/", handlers.SignupHandler(store, userService))
			ts := httptest.NewServer(mux)
			defer ts.Close()

			// request
			client := &http.Client{
				CheckRedirect: func(*http.Request, []*http.Request) error {
					return http.ErrUseLastResponse // do not follow redirects
				},
			}

			resp, err := client.PostForm(ts.URL+"/signup/"+code, url.Values{"password": {password}, "confirmation": {password}})
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			// ensure status code 302
			if resp.StatusCode != http.StatusFound {
				t.Fatalf("expected status code %d but got %d\n", http.StatusFound, resp.StatusCode)
			}

			// ensure redirect to protected area
			location := resp.Header.Get("Location")
			if location != handlers.ProtectedAreaPublicHome {
				t.Fatalf("expected redirect to %s but was to %s\n", handlers.ProtectedAreaPublicHome, location)
			}

			// ensure cookie has been set
			sessionCookieFound := false
			for _, cookie := range resp.Cookies() {
				if cookie.Name == handlers.SessionName {
					sessionCookieFound = true
					break
				}
			}
			if !sessionCookieFound {
				t.Fatalf("expected %s cookie to exist but didn't: %s\n", handlers.SessionName, resp.Cookies())
			}

			// ensure hash has been stored and code deleted
			var (
				storedHash sql.NullString
				storedCode string
			)
			row := db.QueryRow("SELECT hash, code FROM users WHERE email = $1", email)
			if err := row.Scan(&storedHash, &storedCode); err != nil {
				t.Fatal(err)
			}

			if !storedHash.Valid {
				t.Fatal("expected hash not to be nil")
			}

			if storedCode != "" {
				t.Fatal("expected code to be empty")
			}
		},
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

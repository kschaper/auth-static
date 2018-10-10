package handlers_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/kschaper/auth-static/handlers"
	"github.com/kschaper/auth-static/services"
)

func TestSigninFormHandler(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"success": func(t *testing.T) {
			// server
			store := sessions.NewCookieStore([]byte("abc"))
			mux := http.NewServeMux()
			mux.HandleFunc("/signin/", handlers.SigninFormHandler(store))
			ts := httptest.NewServer(mux)
			defer ts.Close()

			// request
			req, err := http.Get(ts.URL + "/signin")
			if err != nil {
				t.Fatal(err)
			}
			defer req.Body.Close()

			// ensure status code 200
			if req.StatusCode != http.StatusOK {
				t.Fatalf("expected status code %d but got %d\n", http.StatusOK, req.StatusCode)
			}

			// ensure HTML is correct
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			html := string(body)
			expected := `action="/signin"`
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

func TestSigninHandler(t *testing.T) {
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
				t.Fatal(err)
			}

			// get user id
			id, err := userService.GetIDByCode(code)
			if err != nil {
				t.Fatal(err)
			}

			// set password
			err = userService.UpdatePassword(id, password, password)
			if err != nil {
				t.Fatal(err)
			}

			// server
			store := sessions.NewCookieStore([]byte("abc"))
			mux := http.NewServeMux()
			mux.HandleFunc("/signin/", handlers.SigninHandler(store, userService))
			ts := httptest.NewServer(mux)
			defer ts.Close()

			// request
			client := &http.Client{
				CheckRedirect: func(*http.Request, []*http.Request) error {
					return http.ErrUseLastResponse // do not follow redirects
				},
			}

			resp, err := client.PostForm(ts.URL+"/signin/", url.Values{"password": {password}, "email": {email}})
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
		},
		"fail": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
			)

			// server
			store := sessions.NewCookieStore([]byte("abc"))
			mux := http.NewServeMux()
			mux.HandleFunc("/signin/", handlers.SigninHandler(store, userService))
			ts := httptest.NewServer(mux)
			defer ts.Close()

			// request
			client := &http.Client{
				CheckRedirect: func(*http.Request, []*http.Request) error {
					return http.ErrUseLastResponse // do not follow redirects
				},
			}

			resp, err := client.PostForm(ts.URL+"/signin/", url.Values{"password": {"xxx"}, "email": {"xxx@example.com"}})
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
			expected := "/signin"
			if location != expected {
				t.Fatalf("expected redirect to %s but was to %s\n", expected, location)
			}
		},
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

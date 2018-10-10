package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kschaper/auth-static/handlers"
	"github.com/kschaper/auth-static/services"

	"github.com/gorilla/sessions"
	uuid "github.com/satori/go.uuid"
)

func TestAuthenticationHandler(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"authenticated": func(t *testing.T) {
			var (
				db           = db(t)
				userService  = &services.UserService{DB: db}
				requestPath  = "/private/secret.jpg"
				redirectPath = "/internal/secret.jpg"
				contentType  = "image/jpeg"
			)

			// create user
			code, err := userService.Create("webmaster@example.com")
			if err != nil {
				t.Fatal(err)
			}

			// handler
			store := sessions.NewCookieStore([]byte("abc"))
			handler := handlers.AuthenticationHandler(store, userService)
			w := httptest.NewRecorder()

			// request
			req, err := http.NewRequest("GET", requestPath, nil)
			if err != nil {
				t.Fatal(err)
			}

			// get the session
			session, err := store.Get(req, handlers.SessionName)
			if err != nil {
				t.Fatal(err)
			}

			// get user id
			id, err := userService.GetIDByCode(code)
			if err != nil {
				t.Fatal(err)
			}

			// put the user id in session
			session.Values[handlers.UserIDKey] = id.String()
			if err := session.Save(req, w); err != nil {
				t.Fatal(err)
			}

			// get the session cookie: `sessionName=value==`
			cookie := strings.Split(w.Header().Get("Set-Cookie"), ";")[0]

			// set the session cookie for the request
			req.Header.Set("Set-Cookie", cookie)

			// invoke handler
			handler(w, req)

			// ensure status code 200
			if w.Code != http.StatusOK {
				t.Fatalf("expected status code %d but got %d\n", http.StatusOK, w.Code)
			}

			// ensure X-Accel-Redirect header has correct value
			redirectHeader := w.Header().Get("X-Accel-Redirect")
			if redirectHeader != redirectPath {
				t.Fatalf("expected X-Accel-Redirect with path %q but got %q\n", redirectPath, redirectHeader)
			}

			// ensure Content-Type header is correct
			contentTypeHeader := w.Header().Get("Content-Type")
			if contentTypeHeader != contentType {
				t.Fatalf("expected Content-Type with %q but got %q\n", contentType, contentTypeHeader)
			}
		},
		"unauthenticated": func(t *testing.T) {
			var (
				db          = db(t)
				userService = &services.UserService{DB: db}
			)

			// handler
			store := sessions.NewCookieStore([]byte("abc"))
			handler := handlers.AuthenticationHandler(store, userService)
			w := httptest.NewRecorder()

			// request
			req, err := http.NewRequest("GET", "/private/whatever.html", nil)
			if err != nil {
				t.Fatal(err)
			}

			// get the session
			session, err := store.Get(req, handlers.SessionName)
			if err != nil {
				t.Fatal(err)
			}

			// put unkown user id in session
			session.Values[handlers.UserIDKey] = uuid.NewV4().String()
			if err := session.Save(req, w); err != nil {
				t.Fatal(err)
			}

			// get the session cookie: `sessionName=value==`
			cookie := strings.Split(w.Header().Get("Set-Cookie"), ";")[0]

			// set the session cookie for the request
			req.Header.Set("Set-Cookie", cookie)

			// invoke handler
			handler(w, req)

			// ensure status code 404
			if w.Code != http.StatusNotFound {
				t.Fatalf("expected status code %d but got %d\n", http.StatusNotFound, w.Code)
			}

			// ensure X-Accel-Redirect header is not set
			redirectHeader := w.Header().Get("X-Accel-Redirect")
			if redirectHeader != "" {
				t.Fatalf("expected X-Accel-Redirect not to be set but got value %q\n", redirectHeader)
			}
		},
	}

	for n, c := range cases {
		t.Run(n, c)
	}
}

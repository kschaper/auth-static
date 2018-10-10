package handlers

import (
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/satori/go.uuid"

	"github.com/gorilla/sessions"
	"github.com/kschaper/auth-static/services"
)

// AuthenticationHandler gets the user_id from the session and checks if there's a corresponding user in the database.
func AuthenticationHandler(store *sessions.CookieStore, userService *services.UserService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		notFoundText := fmt.Sprintf("%d %s", http.StatusNotFound, http.StatusText(http.StatusNotFound))

		// get session
		session, err := store.Get(r, SessionName)
		if err != nil {
			http.Error(w, notFoundText, http.StatusNotFound)
			return
		}

		// get user_id from session and convert into UUID
		userID := session.Values[UserIDKey]
		if userID == nil {
			http.Error(w, notFoundText, http.StatusNotFound)
			return
		}
		userUUID, err := uuid.FromString(fmt.Sprintf("%s", userID))
		if err != nil {
			http.Error(w, notFoundText, http.StatusNotFound)
			return
		}

		// check if user exists
		if exists, err := userService.Exists(userUUID); !exists || err != nil {
			http.Error(w, notFoundText, http.StatusNotFound)
			return
		}

		// set Content-Type header
		if mime := mime.TypeByExtension(path.Ext(r.URL.Path)); mime != "" {
			w.Header().Set("Content-Type", mime)
		}

		// set header
		w.Header().Set("X-Accel-Redirect", strings.Replace(r.URL.String(), ProtectedAreaDirExternal, ProtectedAreaDirInternal, 1))
	}
}

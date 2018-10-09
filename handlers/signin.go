package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"gitlab.com/kschaper/auth-static/services"
)

type signinFormTplData struct {
	Errors []string // from flash messages
}

const signinFormTpl = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>sign in</title>
  </head>
  <body>
		<h1>sign in</h1>
    <form action="/signin" method="post">
      email: <input type="text" name="email">
      password: <input type="password" name="password">
      <input type="submit" value="sign in">
		</form>
		{{if .Errors}}
			<ul>
				{{range .Errors}}
					<li>{{.}}</li>
				{{end}}
			</ul>
		{{end}}
  </body>
</html>
`

// SigninFormHandler shows the signin form.
func SigninFormHandler(store *sessions.CookieStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl := template.Must(template.New("signin").Parse(signinFormTpl))

		// get session
		session, err := store.Get(r, SessionName)
		if err != nil {
			log.Print(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// template data
		data := signinFormTplData{}
		if flashes := session.Flashes(); len(flashes) > 0 {
			for _, flash := range flashes {
				data.Errors = append(data.Errors, fmt.Sprintf("%s", flash))
			}
		}

		if err := session.Save(r, w); err != nil {
			log.Print(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// show page
		tpl.Execute(w, data)
	}
}

// SigninHandler authenticates and redirects.
func SigninHandler(store *sessions.CookieStore, userService *services.UserService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			email    = r.PostFormValue("email")
			password = r.PostFormValue("password")
		)

		// authenticate
		authenticated, err := userService.Authenticate(email, password)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// get session
		session, err := store.Get(r, SessionName)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// handle authentication failed
		if !authenticated {
			session.AddFlash("email and/or password wrong")
			if err := session.Save(r, w); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}

		// get user id
		id, err := userService.GetIDByEmail(email)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// store user id in session
		session.Values[UserIDKey] = id.String()
		if err := session.Save(r, w); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// redirect to protected area
		http.Redirect(w, r, ProtectedAreaPublicHome, http.StatusFound)
	}
}

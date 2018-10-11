package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"

	"github.com/satori/go.uuid"

	"github.com/gorilla/sessions"
	"github.com/kschaper/auth-static/config"
	"github.com/kschaper/auth-static/services"
)

type signupFormTplData struct {
	Code           string   // from URL
	PasswordMinLen int      // from package services
	Errors         []string // from flash messages
}

const signupFormTpl = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>sign up</title>
  </head>
  <body>
		<h1>sign up</h1>
		<p>The password must have at least {{.PasswordMinLen}} characters.</p>
    <form action="/signup/{{.Code}}" method="post">
      password: <input type="password" name="password">
      again: <input type="password" name="confirmation">
      <input type="submit" value="sign up">
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

// SignupFormHandler shows the signup form.
func SignupFormHandler(conf *config.Config, store *sessions.CookieStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			reg  = regexp.MustCompile("[a-z0-9]{32}")
			code = reg.FindString(r.URL.String()) // TODO: use Gorilla Mux's path vars
			tpl  = template.Must(template.New("signup").Parse(signupFormTpl))
		)

		// get session
		session, err := store.Get(r, conf.SessionName)
		if err != nil {
			log.Print(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// template data
		data := signupFormTplData{
			Code:           code,
			PasswordMinLen: services.PasswordMinLen,
		}

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

// SignupHandler sets the password and redirects.
func SignupHandler(conf *config.Config, store *sessions.CookieStore, userService *services.UserService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			reg          = regexp.MustCompile("[a-z0-9]{32}")
			code         = reg.FindString(r.URL.String()) // TODO: use Gorilla Mux's path vars
			password     = r.PostFormValue("password")
			confirmation = r.PostFormValue("confirmation")
			id           uuid.UUID
		)

		// get session
		session, err := store.Get(r, conf.SessionName)

		// get user id
		if err == nil {
			id, err = userService.GetIDByCode(code)
		}

		// update password
		if err == nil {
			err = userService.UpdatePassword(id, password, confirmation)
		}

		// handle errors
		if err != nil {
			log.Print(err)
			switch err.(type) {
			case services.Error:
				session.AddFlash(err.Error())
				if err := session.Save(r, w); err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/signup/"+code, http.StatusFound)
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// store user id in session
		session.Values[conf.UserIDKey] = id.String()
		if err := session.Save(r, w); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// redirect to protected area
		http.Redirect(w, r, conf.ProtectedAreaDirExternal+conf.ProtectedAreaHome, http.StatusFound)
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gitlab.com/kschaper/auth-static/handlers"
	"gitlab.com/kschaper/auth-static/services"
)

var (
	dsn       = flag.String("dsn", "prod.db", "data source name")
	hashKey   = flag.String("hashkey", "", "cookie authentication key")
	blockKey  = flag.String("blockkey", "", "cookie encryption key")
	keylength = 32
	secure    = flag.Bool("secure", false, "cookie secure flag")
)

// TODO: introduce -host flag
// TODO: introduce -port flag

func main() {
	flag.Parse()

	// validate flags
	if len(*hashKey) != keylength || len(*blockKey) != keylength {
		panic(fmt.Sprintf("please provide hashkey and blockkey both with %d chars", keylength))
	}

	// database client
	client := services.DatabaseClient{DSN: *dsn}
	db, err := client.Open()
	if err != nil {
		panic(err)
	}

	// session
	store := sessions.NewCookieStore([]byte(*hashKey), []byte(*blockKey))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   *secure,
	}

	// services
	userService := &services.UserService{DB: db}

	// routes
	r := mux.NewRouter()
	r.HandleFunc("/signup/{code:[a-z0-9]{32}}", handlers.SignupFormHandler(store)).Methods("GET")
	r.HandleFunc("/signup/{code:[a-z0-9]{32}}", handlers.SignupHandler(store, userService)).Methods("POST")
	r.HandleFunc("/signin", handlers.SigninFormHandler(store)).Methods("GET")
	r.HandleFunc("/signin", handlers.SigninHandler(store, userService)).Methods("POST")
	r.PathPrefix(handlers.ProtectedAreaDirExternal).HandlerFunc(handlers.AuthenticationHandler(store, userService))
	http.Handle("/", r)

	// server
	log.Fatal(http.ListenAndServe("localhost:9000", nil))
}

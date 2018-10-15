package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/kschaper/auth-static/config"
	"github.com/kschaper/auth-static/handlers"
	"github.com/kschaper/auth-static/services"
)

var (
	// web server
	host = flag.String("host", "localhost", "the host the server listens to")
	port = flag.Int("port", 9000, "the port the server listens to")

	// database
	dsn = flag.String("dsn", "prod.db", "data source name")

	// cookie
	hashKey   = flag.String("hashkey", "", "cookie authentication key")
	blockKey  = flag.String("blockkey", "", "cookie encryption key")
	keylength = 32
	secure    = flag.Bool("secure", false, "cookie secure flag")

	// paths
	external = flag.String("external", "/private/", "protected area external dir")
	internal = flag.String("internal", "/internal/", "protected area internal dir")
	home     = flag.String("home", "main.html", "protected area home, default: main.html")
)

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

	// config
	conf := config.NewConfig()
	conf.ProtectedAreaDirExternal = *external
	conf.ProtectedAreaDirInternal = *internal
	conf.ProtectedAreaHome = *home

	// routes
	r := mux.NewRouter()
	r.HandleFunc("/signup/{code:[a-z0-9]{32}}", handlers.SignupFormHandler(conf, store)).Methods("GET")
	r.HandleFunc("/signup/{code:[a-z0-9]{32}}", handlers.SignupHandler(conf, store, userService)).Methods("POST")
	r.HandleFunc("/signin", handlers.SigninFormHandler(conf, store)).Methods("GET")
	r.HandleFunc("/signin", handlers.SigninHandler(conf, store, userService)).Methods("POST")
	r.PathPrefix(conf.ProtectedAreaDirExternal).HandlerFunc(handlers.AuthenticationHandler(conf, store, userService))
	http.Handle("/", r)

	// server
	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("Server running at http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

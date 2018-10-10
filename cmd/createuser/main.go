package main

import (
	"flag"
	"fmt"

	"github.com/kschaper/auth-static/services"
	_ "github.com/mattn/go-sqlite3"
)

var (
	email = flag.String("email", "", "email of new user")
	dsn   = flag.String("dsn", "prod.db", "data source name")
	usage = "createuser -email <email> -dsn <dsn>"
)

func main() {
	flag.Parse()

	if *email == "" {
		fmt.Printf("error: no email given\n%s\n", usage)
		return
	}

	if *dsn == "" {
		fmt.Printf("error: no dsn given\n%s\n", usage)
		return
	}

	client := &services.DatabaseClient{DSN: *dsn}
	db, err := client.Open()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	userService := &services.UserService{DB: db}
	code, err := userService.Create(*email)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	fmt.Printf("successfully saved user with email %q and code %q\n", *email, code)
}

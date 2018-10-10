package main

import (
	"flag"
	"fmt"

	"github.com/kschaper/auth-static/services"
	_ "github.com/mattn/go-sqlite3"
)

var dsn = flag.String("dsn", "prod.db", "data source name")

func main() {
	flag.Parse()
	fmt.Printf("init db with dsn %q\n", *dsn)

	client := services.DatabaseClient{DSN: *dsn}
	if _, err := client.Open(); err != nil {
		panic(err)
	}

	fmt.Println("init db successful")
}

package main

import (
	"encoding/hex"
	"fmt"

	"github.com/gorilla/securecookie"
)

func gen() string {
	key := securecookie.GenerateRandomKey(16)
	return hex.EncodeToString(key)
}

func main() {
	fmt.Printf("%s\n%s\n", gen(), gen())
}

package main

import (
	"encoding/base64"
	"fmt"

	"github.com/gorilla/securecookie"
)

func main() {
	enc := base64.StdEncoding
	hashKey := securecookie.GenerateRandomKey(64)
	blockKey := securecookie.GenerateRandomKey(32)

	fmt.Printf("hash_key = \"%s\"\n", enc.EncodeToString(hashKey))
	fmt.Printf("block_key = \"%s\"\n", enc.EncodeToString(blockKey))
}

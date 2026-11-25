package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// The hash-password command prints the bcrypt hash to put in OWNER_PASSWORD_HASH. It reads the password from
// standard input so it never lands in the shell history:
//
//	read -rs PASSWORD && printf '%s' "$PASSWORD" | go run ./cmd/hash-password; unset PASSWORD
func main() {
	password, readError := bufio.NewReader(os.Stdin).ReadString('\n')
	if readError != nil && password == "" {
		log.Fatalf("read the password from standard input: %v", readError)
	}
	password = strings.TrimRight(password, "\r\n")
	if password == "" {
		log.Fatal("the password is empty")
	}

	passwordHash, hashError := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashError != nil {
		log.Fatalf("hash the password: %v", hashError)
	}

	fmt.Println(string(passwordHash))
}

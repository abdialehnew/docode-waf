package main

import (
"fmt"
"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "Admin123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Bcrypt Hash: %s\n", string(hashedPassword))
}

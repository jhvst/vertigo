package main

import (
	"code.google.com/p/go.crypto/bcrypt"
)

// Generates bcrypt hash from plaintext password
func GenerateHash(password string) []byte {
	hex := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(hex, 10)
	if err != nil {
		panic(err)
	}
	return hashedPassword
}

// Compares bcrypt password with a plaintext one. Returns true if passwords match
// and false if they do not.
func CompareHash(digest []byte, password string) bool {
	hex := []byte(password)
	if err := bcrypt.CompareHashAndPassword(digest, hex); err == nil {
		return true
	}
	return false
}

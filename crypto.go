package vertigo

// This file contains two cryptographic functions for both storing and comparing passwords.
// You should not modify this file unless you know what you are doing.

import (
	"code.google.com/p/go.crypto/bcrypt"
)

// GenerateHash generates bcrypt hash from plaintext password
func GenerateHash(password string) []byte {
	hex := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(hex, 10)
	if err != nil {
		panic(err)
	}
	return hashedPassword
}

// CompareHash compares bcrypt password with a plaintext one. Returns true if passwords match
// and false if they do not.
func CompareHash(digest []byte, password string) bool {
	hex := []byte(password)
	if err := bcrypt.CompareHashAndPassword(digest, hex); err == nil {
		return true
	}
	return false
}

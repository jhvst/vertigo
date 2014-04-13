package main

import (
	"code.google.com/p/go.crypto/bcrypt"
)

func GenerateHash(password string) []byte {
	hex := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(hex, 10)
	if err != nil {
		panic(err)
	}
	return hashedPassword
}

func CompareHash(digest []byte, password string) bool {
	hex := []byte(password)
	if err := bcrypt.CompareHashAndPassword(digest, hex); err == nil {
		return true
	}
	return false
}

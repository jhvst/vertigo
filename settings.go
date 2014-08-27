package main

import (
	"io"
	"runtime"

	"code.google.com/p/go-uuid/uuid"
)

func init() {
	settings, err := SessionCookie()
	if err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// Creates a session cookie
func SessionCookie() (Vertigo, error) {
	var settings Vertigo
	w, err := sugarcane.Open("settings.vtg")
	if err != nil {
		return "", err
	}
	data, err := sugarcane.Read("settings.vtg")
	if err != nil {
		return "", err
	}
	err = sugarcane.Scan(&settings, data)
	if err == io.EOF {
		settings.CookieHash = uuid.New()
		sugarcane.Insert(settings, w)
		return SessionCookie()
	}
	if err != nil {
		return "", err
	}
	return settings, nil
}

type Vertigo struct {
	Hostname    string
	CookieHash  string
	Description string
	Mailer      MailgunSettings
}

type MailgunSettings struct {
	Domain     string
	PrivateKey string
	PublicKey  string
}

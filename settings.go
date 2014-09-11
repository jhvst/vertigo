package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"code.google.com/p/go-uuid/uuid"
	"github.com/martini-contrib/render"
)

type Vertigo struct {
	Name        string          `json:"name" form:"name"`
	Hostname    string          `json:"hostname" form:"hostname" binding:"required"`
	Firstrun    bool            `json:"firstrun"`
	CookieHash  string          `json:"cookiehash"`
	Description string          `json:"description" form:"description"`
	Mailer      MailgunSettings `json:"mailgun"`
}

type MailgunSettings struct {
	Domain     string `json:"domain" form:"mgdomain"`
	PrivateKey string `json:"key" form:"mgprikey"`
	PublicKey  string `json:"pubkey" form:"mgpubkey"`
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var cookie *string = flag.String("cookie", SessionCookie(), "session cookie used to handle logins etc")
var firstrun *bool = flag.Bool("firstrun", Firstrun(), "checks whether the installation is new and needs settings wizard to be shown")

// Firstrun is a flag flag shorthand function which checks whether the application has been started for the first time
// and whether the installation wizard should be called when accessing homepage.
func Firstrun() bool {
	var settings Vertigo
	data, err := ioutil.ReadFile("settings.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		panic(err)
	}
	return settings.Firstrun
}

// Returns a session cookie. Creates the whole settings file if it does not already exist.
func SessionCookie() string {
	var settings Vertigo
	_, err := os.OpenFile("settings.json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadFile("settings.json")
	if err != nil {
		panic(err)
	}

	// If settings file is empty, we presume its a first run.
	if len(data) == 0 {
		settings.CookieHash = uuid.New()
		settings.Firstrun = true
		jsonconfig, err := json.Marshal(settings)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile("settings.json", jsonconfig, 0600)
		if err != nil {
			panic(err)
		}
		return SessionCookie()
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		panic(err)
	}
	return settings.CookieHash
}

// UpdateSettings is a route which updates the local .json settings file.
// It is supposed to be disabled after the first run. Therefore the JSON route is not available for now.
func UpdateSettings(res render.Render, settings Vertigo) {
	if *firstrun == false {
		log.Println("Somebody tried to change your local settings...")
		res.JSON(406, map[string]interface{}{"error": "You are not allowed to change underlying settings this time."})
		return
	}
	settings.CookieHash = *cookie
	settings.Firstrun = false
	err := flag.Set("firstrun", "false")
	if err != nil {
		panic(err)
	}
	jsonconfig, err := json.Marshal(settings)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	err = ioutil.WriteFile("settings.json", jsonconfig, 0600)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	res.Redirect("/user/register", 302)
}

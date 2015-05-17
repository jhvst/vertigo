package main

import (
	"log"
	"net/http"

	. "github.com/9uuso/vertigo/databases/gorm"
	. "github.com/9uuso/vertigo/misc"
	. "github.com/9uuso/vertigo/settings"

	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
)

// ReadSettings is a route which reads the local settings.json file.
func ReadSettings(req *http.Request, res render.Render, s sessions.Session) {
	var safesettings Vertigo
	safesettings = *Settings
	safesettings.CookieHash = ""
	switch Root(req) {
	case "api":
		res.JSON(200, safesettings)
		return
	case "user":
		res.HTML(200, "settings", safesettings)
		return
	}
}

// UpdateSettings is a route which updates the local .json settings file.
func UpdateSettings(req *http.Request, res render.Render, settings Vertigo, s sessions.Session) {
	if Settings.Firstrun == false {
		var user User
		user, err := user.Session(s)
		if err != nil {
			log.Println(err)
			res.JSON(406, map[string]interface{}{"error": "You are not allowed to change the settings this time."})
			return
		}
		settings.CookieHash = Settings.CookieHash
		settings.Firstrun = Settings.Firstrun
		err = settings.Save()
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}
		switch Root(req) {
		case "api":
			res.JSON(200, map[string]interface{}{"success": "Settings were successfully saved"})
			return
		case "user":
			res.Redirect("/user", 302)
			return
		}
	}
	settings.Firstrun = false
	settings.AllowRegistrations = true
	err := settings.Save()
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch Root(req) {
	case "api":
		res.JSON(200, map[string]interface{}{"success": "Settings were successfully saved"})
		return
	case "user":
		res.Redirect("/user/register", 302)
		return
	}
}

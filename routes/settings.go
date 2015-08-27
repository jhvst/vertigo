package routes

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	. "github.com/9uuso/vertigo/databases/sqlx"
	. "github.com/9uuso/vertigo/misc"
	"github.com/9uuso/vertigo/render"
	. "github.com/9uuso/vertigo/settings"

	"github.com/martini-contrib/sessions"
)

// ReadSettings is a route which reads the local settings.json file.
func ReadSettings(w http.ResponseWriter, r *http.Request, s sessions.Session) {
	var safesettings Vertigo
	safesettings = *Settings
	safesettings.CookieHash = ""
	switch Root(r) {
	case "api":
		render.R.JSON(w, 200, safesettings)
		return
	case "user":
		render.R.HTML(w, 200, "settings", safesettings)
		return
	}
}

// UpdateSettings is a route which updates the local .json settings file.
func UpdateSettings(w http.ResponseWriter, r *http.Request, settings Vertigo, s sessions.Session) {
	if Settings.Firstrun == false {
		var user User
		user, err := user.Session(s)
		if err != nil {
			log.Println(err)
			render.R.JSON(w, 406, map[string]interface{}{"error": "You are not allowed to change the settings this time."})
			return
		}
		settings.CookieHash = Settings.CookieHash
		settings.Firstrun = Settings.Firstrun
		err = settings.Save()
		if err != nil {
			log.Println(err)
			render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
			return
		}
		switch Root(r) {
		case "api":
			render.R.JSON(w, 200, map[string]interface{}{"success": "Settings were successfully saved"})
			return
		case "user":
			http.Redirect(w, r, "/user", 302)
			return
		}
	}
	settings.Hostname = strings.TrimRight(settings.Hostname, "/")
	u, err := url.Parse(settings.Hostname)
	if err != nil {
		log.Println(err)
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	settings.URL = *u
	settings.Firstrun = false
	settings.AllowRegistrations = true
	err = settings.Save()
	if err != nil {
		log.Println(err)
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch Root(r) {
	case "api":
		render.R.JSON(w, 200, map[string]interface{}{"success": "Settings were successfully saved"})
		return
	case "user":
		http.Redirect(w, r, "/user/register", 302)
		return
	}
}

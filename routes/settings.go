package routes

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	. "github.com/9uuso/vertigo/databases/sqlx"
	"github.com/9uuso/vertigo/render"
	. "github.com/9uuso/vertigo/session"

	"github.com/gorilla/context"
)

func GetSettings(r *http.Request) (Vertigo, error) {
	rv, ok := context.GetOk(r, "settings")
	if !ok {
		return Vertigo{}, errors.New("context not set")
	}
	return rv.(Vertigo), nil
}

// ReadSettings is a route which reads the local settings.json file.
func ReadSettings(w http.ResponseWriter, r *http.Request) {
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
func UpdateSettings(w http.ResponseWriter, r *http.Request) {

	settings, err := GetSettings(r)
	if err != nil {
		log.Println("route UpdateSettings, settings context:", err)
		render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	if Settings.Firstrun {

		settings.Hostname = strings.TrimRight(settings.Hostname, "/")
		_, err := url.Parse(settings.Hostname)
		if err != nil {
			log.Println("route UpdateSettings, url.Parse:", err)
			render.R.JSON(w, 500, map[string]interface{}{"error": "Internal server error"})
			return
		}
		settings.AllowRegistrations = true

		Settings, err = settings.Insert()
		if err != nil {
			log.Println("route UpdateSettings, settings.Save:", err)
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

	_, ok := SessionGetValue(r, "id")
	if !ok {
		log.Println("route UpdateSettings, SessionGetValue:", ok)
		render.R.JSON(w, 401, map[string]interface{}{"error": "Unauthorized"})
		return
	}

	Settings, err = settings.Update()
	if err != nil {
		log.Println("route UpdateSettings, firstrun settings.Save:", err)
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

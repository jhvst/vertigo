// Package settings defines site-wide settings variable, which is stored in
// local disk in JSON format. The file is written to disk with 0600 permissions
// and is called settings.json.
package settings

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/pborman/uuid"
)

// Vertigo struct is used as a site wide settings structure.
// Firstrun and CookieHash are generated and controlled by the application and should not be
// rendered or made editable anywhere on the site.
type Vertigo struct {
	Name               string  `json:"name" form:"name" binding:"required"`
	Hostname           string  `json:"hostname" form:"hostname" binding:"required"`
	URL                url.URL `json:"-,omitempty"`
	Firstrun           bool    `json:"firstrun,omitempty"`
	CookieHash         string  `json:"cookiehash,omitempty"`
	AllowRegistrations bool    `json:"allowregistrations" form:"allowregistrations"`
	Description        string  `json:"description" form:"description" binding:"required"`
	Mailer             SMTP    `json:"smtp"`
}

// SMTP holds information necessary to send account recovery email.
type SMTP struct {
	Login    string `json:"login" form:"login"`
	Port     int    `json:"port" form:"port"`
	Password string `json:"password" form:"password"`
	Hostname string `json:"hostname" form:"smtp-hostname"`
}

/*

Settings is the global variable which holds settings stored in the settings.json file.
After importing the package, you can manipulate the variable using the Settings keyword.
Although, as mentioned in the Vertigo struct, be careful when dealing with the Firstrun and CookieHash values.
Rewriting Firstrun to true will render installation wizard on homepage, letting anyone redeclare your settings.
CookieHash on the other hand will return the secret hash used to sign your cookies, which might result in accounts getting compromised.

	package mypackage

	import (
		"fmt"

		"github.com/9uuso/vertigo/settings"
	)

	func Foobar() {
		fmt.Println(Settings.Name)
		// Output: Foobar's Blog
		settings = *Settings
		settings.Name = "Juuso's Blog"
		settings.Save()
		if err != nil {
			panic(err)
		}
		fmt.Println(Settings.Name)
		// Output: Juuso's Blog
	}

	func ReadSettings() {
		var safesettings Vertigo
		safesettings = *Settings
		safesettings.CookieHash = "" // CookieHash wont be exposed
		switch Root(req) {
		case "api":
			render.JSON(200, safesettings)
			return
		case "user":
			render.HTML(200, "settings", safesettings)
			return
		}
	}
	
*/
var Settings = VertigoSettings()

// VertigoSettings populates the global namespace with data from settings.json.
// If the file does not exist, it creates it.
func VertigoSettings() *Vertigo {
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
		var settings Vertigo
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
		return VertigoSettings()
	}

	var settings *Vertigo
	if err := json.Unmarshal(data, &settings); err != nil {
		panic(err)
	}
	return settings
}

// Save or Settings.Save is a method which replaces the global Settings structure with the structure is is called with.
// It has builtin variable declaration which prevents you from overwriting CookieHash field.
func (settings *Vertigo) Save() error {
	var old Vertigo
	data, err := ioutil.ReadFile("settings.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &old); err != nil {
		return err
	}
	Settings = settings
	settings.CookieHash = old.CookieHash // this to assure that cookiehash cannot be overwritten even if system is hacked
	jsonconfig, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("settings.json", jsonconfig, 0600)
	if err != nil {
		return err
	}
	return nil
}

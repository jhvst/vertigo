// Settings.go includes everything you would think site-wide settings need. It also contains a installation wizard
// route at the bottom of the file. You generally should not need to change anything in here.
package settings

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"

	"code.google.com/p/go-uuid/uuid"
)

// Vertigo struct is used as a site wide settings structure. Different from posts and person
// it is saved on local disk in JSON format.
// Firstrun and CookieHash are generated and controlled by the application and should not be
// rendered or made editable anywhere on the site.
type Vertigo struct {
	Name               string  `json:"name" form:"name" binding:"required"`
	Hostname           string  `json:"hostname" form:"hostname" binding:"required"`
	URL                url.URL `json:"-"`
	Firstrun           bool    `json:"firstrun,omitempty"`
	CookieHash         string  `json:"cookiehash,omitempty"`
	AllowRegistrations bool    `json:"allowregistrations" form:"allowregistrations"`
	Markdown           bool    `json:"markdown" form:"markdown"`
	Description        string  `json:"description" form:"description" binding:"required"`
	Mailer             SMTP    `json:"smtp"`
}

// MailgunSettings holds the API keys necessary to send account recovery email.
// You can find the necessary values for these structures in https://mailgun.com/cp
type SMTP struct {
	Login    string `json:"login" form:"login"`
	Port     int    `json:"port" form:"port"`
	Password string `json:"password" form:"password"`
	Hostname string `json:"hostname" form:"smtp-hostname"`
}

// Settings is a global variable which holds settings stored in the settings.json file.
// You can call it globally anywhere by simply using the Settings keyword. For example
// fmt.Println(Settings.Name) will print out your site's name.
// As mentioned in the Vertigo struct, be careful when dealing with the Firstun and CookieHash values.
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

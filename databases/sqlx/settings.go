package sqlx

import (
	"log"

	"github.com/pborman/uuid"
)

// Vertigo struct is used as a site wide settings structure.
// Firstrun and CookieHash are generated and controlled by the application and should not be
// rendered or made editable anywhere on the site.
type Vertigo struct {
	ID                 int    `json:"-,omitempty"`
	Name               string `json:"name" form:"name" binding:"required"`
	Hostname           string `json:"hostname" form:"hostname" binding:"required"`
	Firstrun           bool   `json:"firstrun,omitempty"`
	CookieHash         string `json:"cookiehash,omitempty"`
	AllowRegistrations bool   `json:"allowregistrations" form:"allowregistrations"`
	Description        string `json:"description" form:"description" binding:"required"`
	MailerLogin        string `json:"mailerlogin" form:"mailerlogin"`
	MailerPort         int    `json:"mailerport" form:"mailerport"`
	MailerPassword     string `json:"mailerpassword" form:"mailerpassword"`
	MailerHostname     string `json:"mailerhostname" form:"mailerhostname"`
}

/*

Settings is the global variable which holds site-wide settings.
After importing the package, you can manipulate the variable using the Settings keyword.
Although, as mentioned in the Vertigo struct, be careful when dealing with the Firstrun and CookieHash values.
Rewriting Firstrun to true will render installation wizard on homepage, letting anyone redeclare your settings.
CookieHash on the other hand will return the secret hash used to sign your cookies, which might result in accounts getting compromised.

	package mypackage

	import (
		"fmt"

		"github.com/toldjuuso/vertigo/settings"
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

// Insert or settings.Insert inserts Vertigo settings object into database.
// Fills settings.ID, settings.CookieHash and settings.FirstRun automatically.
// Returns *Vertigo and error object.
func (settings Vertigo) Insert() (*Vertigo, error) {
	settings.ID = 1
	settings.CookieHash = uuid.New()
	settings.Firstrun = false
	_, err := db.NamedExec(`INSERT INTO settings (id, name, hostname, firstrun, cookiehash, allowregistrations, description, mailerlogin, mailerport, mailerpassword, mailerhostname)
		VALUES (:id, :name, :hostname, :firstrun, :cookiehash, :allowregistrations, :description, :mailerlogin, :mailerport, :mailerpassword, :mailerhostname)`, settings)
	if err != nil {
		return &settings, err
	}
	return &settings, nil
}

// Get or settings.Get returns settings saved to database.
// Returns Vertigo and error object.
func (settings Vertigo) Get() (Vertigo, error) {
	var v Vertigo
	v.ID = 1
	stmt, err := db.PrepareNamed("SELECT * FROM settings WHERE id = :id")
	if err != nil {
		return v, err
	}
	err = stmt.Get(&v, v)
	if err != nil {
		return v, err
	}
	return v, nil
}

// Update or settings.Update writes changes made to global settings variable into database.
// Returns *Vertigo and an error object.
func (settings Vertigo) Update() (*Vertigo, error) {
	settings.ID = 1
	settings.Firstrun = false
	settings.CookieHash = Settings.CookieHash
	_, err := db.NamedExec(
		"UPDATE settings SET name = :name, hostname = :hostname, firstrun = :firstrun, allowregistrations = :allowregistrations, description = :description, mailerlogin = :mailerlogin, mailerport = :mailerport, mailerpassword = :mailerpassword, mailerhostname = :mailerhostname WHERE id = :id",
		settings)
	if err != nil {
		return &settings, err
	}
	return &settings, nil
}

// VertigoSettings populates the global Settings object with data from database.
// If no records exist, it creates one.
func VertigoSettings() *Vertigo {
	var settings Vertigo
	settings, err := settings.Get()
	if err != nil {
		log.Println("settings are empty")
		// If settings file is empty, we presume its a first run.
		if err.Error() == "sql: no rows in result set" {
			settings.Firstrun = true
			return &settings
		}
		panic(err)
	}
	return &settings
}

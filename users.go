// This file contains about everything related to users aka users. At the top you will find routes
// and at the bottom you can find CRUD options. Some functions in this file are analogous
// to the ones in posts.go.
package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/mailgun/mailgun-go"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	_ "github.com/mattn/go-sqlite3"
)

// User struct holds all relevant data for representing user accounts on Vertigo.
// A complete User struct also includes Posts field (type []Post) which includes
// all posts made by the user.
type User struct {
	ID       int64  `json:"id" gorethink:",omitempty"`
	Name     string `json:"name" form:"name" binding:"required" gorethink:"name"`
	Password string `json:"password,omitempty" form:"password" gorethink:"-,omitempty" sql:"-"`
	Recovery string `json:",omitempty" gorethink:"recovery"`
	Digest   []byte `json:"-" gorethink:"digest"`
	Email    string `json:"email,omitempty" form:"email" binding:"required" gorethink:"email"`
	Posts    []Post `json:"posts" gorethink:"posts"`
}

// CreateUser is a route which creates a new user struct according to posted parameters.
// Requires session cookie.
// Returns created user struct for API requests and redirects to "/user" on frontend ones.
func CreateUser(req *http.Request, res render.Render, db *gorm.DB, s sessions.Session, user User) {
	if user.Unique(db) == false {
		res.JSON(422, map[string]interface{}{"error": "Email already in use"})
		return
	}
	user, err := user.Insert(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	switch root(req) {
	case "api":
		res.JSON(200, user)
		return
	case "user":
		res.Redirect("/user/login", 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// DeleteUser is a route which deletes a user from database according to session cookie.
// The function calls Login function inside, so it also requires password in POST data.
// Currently unavailable function on both API and frontend side.
func DeleteUser(req *http.Request, res render.Render, db *gorm.DB, s sessions.Session, user User) {
	user, err := user.Login(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	err = user.Delete(db, s)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	switch root(req) {
	case "api":
		s.Delete("user")
		res.JSON(200, map[string]interface{}{"status": "User successfully deleted"})
		return
	case "user":
		s.Delete("user")
		res.HTML(200, "User successfully deleted", nil)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// ReadUser is a route which fetches user according to parameter "id" on API side and according to retrieved
// session cookie on frontend side.
// Returns user struct with all posts merged to object on API call. Frontend call will render user "home" page, "user/index.tmpl".
func ReadUser(req *http.Request, params martini.Params, res render.Render, s sessions.Session, db *gorm.DB) {
	var user User
	switch root(req) {
	case "api":
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			panic(err)
		}
		user.ID = int64(id)
		user, err := user.Get(db)
		if err != nil {
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			log.Println(err)
			return
		}
		res.JSON(200, user)
		return
	case "user":
		user, err := user.Session(db, s)
		if err != nil {
			s.Set("user", -1)
			res.HTML(500, "error", err)
			log.Println(err)
			return
		}
		res.HTML(200, "user/index", user)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// ReadUsers is a route only available on API side, which fetches all users with post data merged.
// Returns complete list of users on success.
func ReadUsers(res render.Render, db *gorm.DB) {
	var user User
	users, err := user.GetAll(db)
	if err != nil {
		res.JSON(500, err)
		log.Println(err)
		return
	}
	res.JSON(200, users)
}

// EmailIsUnique returns bool value acoording to whether user email already exists in database with called user struct.
// The function is used to make sure two users do not register under the same email. This limitation could however be removed,
// as by default primary key for tables used by Vertigo is ID, not email.
func (user User) Unique(db *gorm.DB) bool {
	if db.Where(&User{Email: user.Email}).First(&user).RecordNotFound() {
		return true
	}
	return false
}

// LoginUser is a route which compares plaintext password sent with POST request with
// hash stored in database. On successful request returns session cookie named "user", which contains
// user's ID encrypted, which is the primary key used in database table.
// When called by API it responds with user struct without post data merged.
// On frontend call it redirect the client to "/user" page.
func LoginUser(req *http.Request, s sessions.Session, res render.Render, db *gorm.DB, user User) {
	user, err := user.Login(db)
	if err != nil {
		if err.Error() == "wrong username or password" {
			res.HTML(401, "user/login", "Wrong username or password.")
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	switch root(req) {
	case "api":
		s.Set("user", user.ID)
		res.JSON(200, user)
		return
	case "user":
		s.Set("user", user.ID)
		res.Redirect("/user", 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// RecoverUser is a route of the first step of account recovery, which sends out the recovery
// email etc. associated function calls.
func RecoverUser(req *http.Request, s sessions.Session, res render.Render, db *gorm.DB, user User) {
	user, err := user.Recover(db)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch root(req) {
	case "api":
		res.JSON(200, map[string]interface{}{"success": "We've sent you a link to your email which you may use you reset your password."})
		return
	case "user":
		res.Redirect("/user/login", 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// ResetUserPassword is a route which is called when accessing the page generated dispatched with
// account recovery emails.
func ResetUserPassword(req *http.Request, params martini.Params, s sessions.Session, res render.Render, db *gorm.DB, user User) {
	id, err := strconv.Atoi(params["id"])
	user.ID = int64(id)
	entry, err := user.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	if entry.Recovery == params["recovery"] {
		entry.Password = user.Password
		_, err := user.Update(db, entry)
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		}
		switch root(req) {
		case "api":
			res.JSON(200, map[string]interface{}{"success": "Password was updated successfully."})
			return
		case "user":
			res.Redirect("/user/login", 302)
			return
		}
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// LogoutUser is a route which deletes session cookie "user", from the given client.
// On API call responds with HTTP 200 body and on frontend the client is redirected to homepage "/".
func LogoutUser(req *http.Request, s sessions.Session, res render.Render) {
	s.Delete("user")
	switch root(req) {
	case "api":
		res.JSON(200, map[string]interface{}{"success": "You've been logged out."})
		return
	case "user":
		res.Redirect("/", 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// Login or user.Login is a function which retrieves user according to given .Email field.
// The function then compares the retrieved object's .Digest field with given .Password field.
// If the .Password and .Digest match, the function returns the requested User struct, but with
// the .Password and .Digest omitted.
func (user User) Login(db *gorm.DB) (User, error) {
	password := user.Password
	db.Where(&User{Email: user.Email}).First(&user)
	if db.Error != nil {
		log.Println(db.Error)
		return user, db.Error
	}
	if !CompareHash(user.Digest, password) {
		return user, errors.New("wrong username or password")
	}
	return user, nil
}

// Update or user.Update updates data of "entry" parameter with the data received from "user".
// Can only used to update Name and Digest fields because of how user.Get works.
// Currently not used elsewhere than in password Recovery, that's why the Digest generation.
func (user User) Update(db *gorm.DB, entry User) (User, error) {
	digest, err := GenerateHash(entry.Password)
	if err != nil {
		return user, err
	}
	entry.Digest = digest
	db.Where([]int64{user.ID}).Find(&user)
	if db.Error != nil {
		log.Println(db.Error)
		return user, db.Error
	}
	db.Model(&user).Updates(User{Name: entry.Name, Digest: entry.Digest})
	if db.Error != nil {
		log.Println(db.Error)
		return user, db.Error
	}
	return user, nil
}

// Recover or user.Recover is used to recover User's password according to user.Email
// The function will insert user.Recovery field with generated UUID string and dispatch an email
// to the corresponding user.Email address. It will also add TTL to Recovery field.
func (user User) Recover(db *gorm.DB) (User, error) {
	db.Where(&User{Email: user.Email}).First(&user)
	if db.Error != nil {
		log.Println(db.Error)
		return user, db.Error
	}

	err := user.InsertRecoveryHash(db)
	if err != nil {
		log.Println(err)
		return user, err
	}

	user, err = user.Get(db)
	if err != nil {
		log.Println(err)
		return user, err
	}

	err = user.SendRecoverMail()
	if err != nil {
		log.Println(err)
		return user, err
	}

	go user.ExpireRecovery(db, 180*time.Minute)

	return user, nil
}

// ExpireRecovery or user.ExpireRecovery sets a TTL according to t to a recovery hash.
// This function is supposed to be run as goroutine to avoid blocking exection for t.
func (user User) ExpireRecovery(db *gorm.DB, t time.Duration) {
	time.Sleep(t)
	err := user.DeleteRecoveryHash(db)
	if err != nil {
		log.Println(err)
	}
	return
}

// Get or user.Get returns User object according to given .ID
// with post information merged, but without the .Digest and .Email field.
func (user User) Get(db *gorm.DB) (User, error) {
	var posts []Post
	db.Where([]int64{user.ID}).Find(&user)
	db.Where(&Post{Author: user.ID}).Find(&posts)
	user.Posts = posts
	if db.Error != nil {
		log.Println(db.Error)
		return user, db.Error
	}
	return user, nil
}

// Session or user.Session returns User object from client session cookie.
// The returned object has post data merged.
func (user User) Session(db *gorm.DB, s sessions.Session) (User, error) {
	data := s.Get("user")
	id, exists := data.(int64)
	if exists {
		var user User
		user.ID = id
		user, err := user.Get(db)
		if err != nil {
			log.Println(err)
			return user, err
		}
		return user, nil
	}
	return user, errors.New("unauthorized")
}

// Delete or user.Delete deletes the user with given .Id from the database.
func (user User) Delete(db *gorm.DB, s sessions.Session) error {
	user, err := user.Session(db, s)
	if err != nil {
		log.Println(err)
		return err
	}
	db.Delete(&user)
	if db.Error != nil {
		log.Println(db.Error)
		return db.Error
	}
	return nil
}

// Insert or user.Insert inserts a new User struct into the database.
// The function creates .Digest hash from .Password.
func (user User) Insert(db *gorm.DB) (User, error) {
	digest, err := GenerateHash(user.Password)
	if err != nil {
		return user, err
	}
	user.Digest = digest
	db.Create(&user)
	if db.Error != nil {
		log.Println(db.Error)
		return user, db.Error
	}
	return user, nil
}

// GetAll or user.GetAll fetches all users with post data merged from the database.
func (user User) GetAll(db *gorm.DB) ([]User, error) {
	var users []User
	db.Find(&users)
	if db.Error != nil {
		log.Println(db.Error)
		return users, db.Error
	}
	return users, nil
}

func (user User) DeleteRecoveryHash(db *gorm.DB) error {
	// should probably be replaced with user.Update call
	db.Where([]int64{user.ID}).Find(&user)
	db.Model(&user).Updates(User{Recovery: ""})
	return nil
}

func (user User) InsertRecoveryHash(db *gorm.DB) error {
	// should probably be replaced with user.Update call
	db.Where([]int64{user.ID}).Find(&user)
	if db.Error != nil {
		log.Println(db.Error)
		return db.Error
	}
	db.Model(&user).Updates(User{Recovery: uuid.New()})
	if db.Error != nil {
		log.Println(db.Error)
		return db.Error
	}
	return nil
}

// SendRecoverMail or user.SendRecoverMail sends mail with Mailgun with pre-filled email layout.
// See Mailgun example on https://gist.github.com/mbanzon/8179682
func (user User) SendRecoverMail() error {
	gun := mailgun.NewMailgun(Settings.Mailer.Domain, Settings.Mailer.PrivateKey, "")
	m := mailgun.NewMessage("Sender <postmaster@"+Settings.Mailer.Domain+">", "Password reset", "Somebody requested password recovery on this email. You may reset your password trough this link: http://"+Settings.Hostname+"/user/reset/"+string(user.ID)+"/"+user.Recovery, "Recipient <"+user.Email+">")
	if _, _, err := gun.Send(m); err != nil {
		return err
	}
	return nil
}

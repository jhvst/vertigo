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
	ID       int64  `json:"id" gorm:"primary_key:yes"`
	Name     string `json:"name" form:"name"`
	Password string `json:"password,omitempty" form:"password" sql:"-"`
	Recovery string `json:"-"`
	Digest   []byte `json:"-"`
	Email    string `json:"email,omitempty" form:"email" binding:"required" sql:"unique"`
	Posts    []Post `json:"posts"`
}

// CreateUser is a route which creates a new user struct according to posted parameters.
// Requires session cookie.
// Returns created user struct for API requests and redirects to "/user" on frontend ones.
func CreateUser(req *http.Request, res render.Render, db *gorm.DB, s sessions.Session, user User) {
	if Settings.AllowRegistrations == false {
		log.Println("Denied a new registration.")
		switch root(req) {
		case "api":
			res.JSON(403, map[string]interface{}{"error": "New registrations are not allowed at this time."})
			return
		case "user":
			res.HTML(403, "user/login", "New registrations are not allowed at this time.")
			return
		}
	}
	user, err := user.Insert(db)
	if err != nil {
		log.Println(err)
		if err.Error() == "UNIQUE constraint failed: users.email" {
			res.JSON(422, map[string]interface{}{"error": "Email already in use"})
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	user, err = user.Login(db)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch root(req) {
	case "api":
		s.Set("user", user.ID)
		user.Password = ""
		res.JSON(200, user)
		return
	case "user":
		s.Set("user", user.ID)
		res.Redirect("/user", 302)
		return
	}
}

// DeleteUser is a route which deletes a user from database according to session cookie.
// The function calls Login function inside, so it also requires password in POST data.
// Currently unavailable function on both API and frontend side.
// func DeleteUser(req *http.Request, res render.Render, db *gorm.DB, s sessions.Session, user User) {
// 	user, err := user.Login(db)
// 	if err != nil {
// 		log.Println(err)
// 		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
// 		return
// 	}
// 	err = user.Delete(db, s)
// 	if err != nil {
// 		log.Println(err)
// 		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
// 		return
// 	}
// 	switch root(req) {
// 	case "api":
// 		s.Delete("user")
// 		res.JSON(200, map[string]interface{}{"status": "User successfully deleted"})
// 		return
// 	case "user":
// 		s.Delete("user")
// 		res.HTML(200, "User successfully deleted", nil)
// 		return
// 	}
// 	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
// }

// ReadUser is a route which fetches user according to parameter "id" on API side and according to retrieved
// session cookie on frontend side.
// Returns user struct with all posts merged to object on API call. Frontend call will render user "home" page, "user/index.tmpl".
func ReadUser(req *http.Request, params martini.Params, res render.Render, s sessions.Session, db *gorm.DB) {
	var user User
	switch root(req) {
	case "api":
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			log.Println(err)
			res.JSON(400, map[string]interface{}{"error": "The user ID could not be parsed from the request URL."})
			return
		}
		user.ID = int64(id)
		user, err := user.Get(db)
		if err != nil {
			log.Println(err)
			if err.Error() == "not found" {
				res.JSON(404, NotFound())
				return
			}
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}
		res.JSON(200, user)
		return
	case "user":
		user, err := user.Session(db, s)
		if err != nil {
			log.Println(err)
			s.Set("user", -1)
			res.HTML(500, "error", err)
			return
		}
		res.HTML(200, "user/index", user)
		return
	}
}

// ReadUsers is a route only available on API side, which fetches all users with post data merged.
// Returns complete list of users on success.
func ReadUsers(res render.Render, db *gorm.DB) {
	var user User
	users, err := user.GetAll(db)
	if err != nil {
		log.Println(err)
		res.JSON(500, err)
		return
	}
	res.JSON(200, users)
}

// LoginUser is a route which compares plaintext password sent with POST request with
// hash stored in database. On successful request returns session cookie named "user", which contains
// user's ID encrypted, which is the primary key used in database table.
// When called by API it responds with user struct.
// On frontend call it redirects the client to "/user" page.
func LoginUser(req *http.Request, s sessions.Session, res render.Render, db *gorm.DB, user User) {
	switch root(req) {
	case "api":
		user, err := user.Login(db)
		if err != nil {
			log.Println(err)
			if err.Error() == "wrong username or password" {
				res.JSON(401, map[string]interface{}{"error": "Wrong username or password."})
				return
			}
			if err.Error() == "not found" {
				res.JSON(404, map[string]interface{}{"error": "User with that email does not exist."})
				return
			}
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}
		s.Set("user", user.ID)
		user.Password = ""
		res.JSON(200, user)
		return
	case "user":
		user, err := user.Login(db)
		if err != nil {
			log.Println(err)
			if err.Error() == "wrong username or password" {
				res.HTML(401, "user/login", "Wrong username or password.")
				return
			}
			if err.Error() == "not found" {
				res.HTML(404, "user/login", "User with that email does not exist.")
				return
			}
			res.HTML(500, "user/login", "Internal server error. Please try again.")
			return
		}
		s.Set("user", user.ID)
		res.Redirect("/user", 302)
		return
	}
}

// RecoverUser is a route of the first step of account recovery, which sends out the recovery
// email etc. associated function calls.
func RecoverUser(req *http.Request, res render.Render, db *gorm.DB, user User) {
	user, err := user.Recover(db)
	if err != nil {
		log.Println(err)
		if err.Error() == "not found" {
			res.JSON(401, map[string]interface{}{"error": "User with that email does not exist."})
			return
		}
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
}

// ResetUserPassword is a route which is called when accessing the page generated dispatched with
// account recovery emails.
func ResetUserPassword(req *http.Request, params martini.Params, res render.Render, db *gorm.DB, user User) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Println(err)
		res.JSON(400, map[string]interface{}{"error": "User ID could not be parsed from request URL."})
		return
	}
	user.ID = int64(id)
	entry, err := user.Get(db)
	if err != nil {
		log.Println(err)
		if err.Error() == "not found" {
			res.JSON(404, map[string]interface{}{"error": "User with that ID does not exist."})
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	// this ensures that accounts won't be compromised by posting recovery string as empty,
	// which would otherwise result in succesful password reset
	UUID := uuid.Parse(params["recovery"])
	if UUID == nil {
		log.Println("there was a problem trying to verify password reset UUID for", entry.Email)
		res.JSON(400, map[string]interface{}{"error": "Could not parse UUID from the request."})
		return
	}
	if entry.Recovery == params["recovery"] {
		entry.Password = user.Password
		digest, err := GenerateHash(entry.Password)
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}
		entry.Digest = digest
		entry.Recovery = " "
		_, err = user.Update(db, entry)
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
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
}

// Login or user.Login is a function which retrieves user according to given .Email field.
// The function then compares the retrieved object's .Digest field with given .Password field.
// If the .Password and .Digest match, the function returns the requested User struct, but with
// the .Password and .Digest omitted.
func (user User) Login(db *gorm.DB) (User, error) {
	password := user.Password
	user, err := user.GetByEmail(db)
	if err != nil {
		return user, err
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
	query := db.Where(&User{ID: user.ID}).Find(&user).Updates(entry)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			return user, errors.New("not found")
		}
		return user, query.Error
	}
	return user, nil
}

// Recover or user.Recover is used to recover User's password according to user.Email
// The function will insert user.Recovery field with generated UUID string and dispatch an email
// to the corresponding user.Email address. It will also add TTL to Recovery field.
func (user User) Recover(db *gorm.DB) (User, error) {

	user, err := user.GetByEmail(db)
	if err != nil {
		return user, err
	}

	var entry User
	entry.Recovery = uuid.New()
	user, err = user.Update(db, entry)
	if err != nil {
		return user, err
	}

	err = user.SendRecoverMail()
	if err != nil {
		return user, err
	}

	go user.ExpireRecovery(db, 180*time.Minute)

	return user, nil
}

// ExpireRecovery or user.ExpireRecovery sets a TTL according to t to a recovery hash.
// This function is supposed to be run as goroutine to avoid blocking exection for t.
func (user User) ExpireRecovery(db *gorm.DB, t time.Duration) {
	time.Sleep(t)

	var entry User
	entry.Recovery = " "
	_, err := user.Update(db, entry)
	if err != nil {
		log.Println(err)
	}
	return
}

// Get or user.Get returns User object according to given .ID
// with post information merged.
func (user User) Get(db *gorm.DB) (User, error) {
	var posts []Post
	query := db.Where(&User{ID: user.ID}).First(&user)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			return user, errors.New("not found")
		}
		return user, query.Error
	}
	query = db.Order("date desc").Where(&Post{Author: user.ID}).Find(&posts)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			user.Posts = make([]Post, 0)
			return user, nil
		}
		return user, query.Error
	}
	user.Posts = posts
	return user, nil
}

// GetByEmail or user.GetByEmail returns User object according to given .Email
// with post information merged.
func (user User) GetByEmail(db *gorm.DB) (User, error) {
	var posts []Post
	query := db.Where(&User{Email: user.Email}).First(&user)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			return user, errors.New("not found")
		}
		return user, query.Error
	}
	query = db.Where(&Post{Author: user.ID}).Find(&posts)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			user.Posts = make([]Post, 0)
			return user, nil
		}
		return user, query.Error
	}
	user.Posts = posts
	return user, nil
}

// Session or user.Session returns user.ID from client session cookie.
// The returned object has post data merged.
func (user User) Session(db *gorm.DB, s sessions.Session) (User, error) {
	data := s.Get("user")
	id, exists := data.(int64)
	if exists {
		var user User
		user.ID = id
		user, err := user.Get(db)
		if err != nil {
			return user, err
		}
		return user, nil
	}
	return user, errors.New("unauthorized")
}

// Delete or user.Delete deletes the user with given ID from the database.
// func (user User) Delete(db *gorm.DB, s sessions.Session) error {
// 	user, err := user.Session(db, s)
// 	if err != nil {
// 		return err
// 	}
// 	query := db.Delete(&user)
// 	if query.Error != nil {
// 		if query.Error == gorm.RecordNotFound {
// 			return errors.New("not found")
// 		}
// 		return query.Error
// 	}
// 	return nil
// }

// Insert or user.Insert inserts a new User struct into the database.
// The function creates .Digest hash from .Password.
func (user User) Insert(db *gorm.DB) (User, error) {
	digest, err := GenerateHash(user.Password)
	if err != nil {
		return user, err
	}
	user.Digest = digest
	user.Posts = make([]Post, 0)
	query := db.Create(&user)
	if query.Error != nil {
		return user, query.Error
	}
	return user, nil
}

// GetAll or user.GetAll fetches all users with post data merged from the database.
func (user User) GetAll(db *gorm.DB) ([]User, error) {
	var users []User
	query := db.Find(&users)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			users = make([]User, 0)
			return users, nil
		}
		return users, query.Error
	}
	for index, user := range users {
		user, err := user.Get(db)
		if err != nil {
			return users, err
		}
		users[index] = user
	}
	return users, nil
}

// SendRecoverMail or user.SendRecoverMail sends mail with Mailgun with pre-filled email layout.
// See Mailgun example on https://gist.github.com/mbanzon/8179682
func (user User) SendRecoverMail() error {
	gun := mailgun.NewMailgun(Settings.Mailer.Domain, Settings.Mailer.PrivateKey, "")
	id := strconv.Itoa(int(user.ID))
	urlhost := urlHost()

	m := mailgun.NewMessage("Password Reset <postmaster@"+Settings.Mailer.Domain+">", "Password Reset", "Somebody requested password recovery on this email. You may reset your password through this link: "+urlhost+"user/reset/"+id+"/"+user.Recovery, "Recipient <"+user.Email+">")
	if _, _, err := gun.Send(m); err != nil {
		return err
	}
	return nil
}


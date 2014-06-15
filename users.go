package main

// This file contains about everything related to persons aka users. At the top you will find routes
// and at the bottom you can find CRUD options. Some functions in this file are analogous
// to the ones in posts.go.

import (
	"errors"
	r "github.com/dancannon/gorethink"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"log"
	"net/http"
)

// Person struct holds all relevant data for representing user accounts on Vertigo.
// A complete Person struct also includes Posts field (type []Post) which includes
// all posts made by the user.
type Person struct {
	ID       string `json:"id" gorethink:",omitempty"`
	Name     string `json:"name" form:"name" binding:"required" gorethink:"name"`
	Password string `form:"password" json:"password,omitempty" gorethink:"-,omitempty"`
	Digest   []byte `json:"digest,omitempty" gorethink:"digest"`
	Email    string `json:"email,omitempty" form:"email" binding:"required" gorethink:"email"`
	Posts    []Post `json:"posts" gorethink:"posts"`
}

// CreateUser is a route which creates a new person struct according to posted parameters.
// Requires session cookie.
// Returns created user struct for API requests and redirects to "/user" on frontend ones.
func CreateUser(req *http.Request, res render.Render, db *r.Session, s sessions.Session, person Person) {
	if !EmailIsUnique(db, person) {
		res.JSON(422, map[string]interface{}{"error": "Email already in use"})
		return
	}
	user, err := person.Insert(db)
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
func DeleteUser(req *http.Request, res render.Render, db *r.Session, s sessions.Session, person Person) {
	person, err := person.Login(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	err = person.Delete(db, s)
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
func ReadUser(req *http.Request, params martini.Params, res render.Render, s sessions.Session, db *r.Session) {
	var person Person
	switch root(req) {
	case "api":
		person.ID = params["id"]
		user, err := person.Get(db)
		if err != nil {
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			log.Println(err)
			return
		}
		res.JSON(200, user)
		return
	case "user":
		user, err := person.Session(db, s)
		if err != nil {
			s.Delete("user")
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
func ReadUsers(res render.Render, db *r.Session) {
	var person Person
	users, err := person.GetAll(db)
	if err != nil {
		res.JSON(500, err)
		log.Println(err)
		return
	}
	res.JSON(200, users)
}

// EmailIsUnique returns bool value acoording to whether user email already exists in database with called user struct.
// The function is used to make sure two persons do not register under the same email. This limitation could however be removed,
// as by default primary key for tables used by Vertigo is ID, not email.
func EmailIsUnique(db *r.Session, person Person) bool {
	row, err := r.Table("users").Filter(func(user r.RqlTerm) r.RqlTerm {
		return user.Field("email").Eq(person.Email)
	}).RunRow(db)
	if err != nil || !row.IsNil() {
		log.Println(err)
		return false
	}
	return true
}

// LoginUser is a route which compares plaintext password sent with POST request with
// hash stored in database. On successful request returns session cookie named "user", which contains
// user's ID encrypted, which is the primary key used in database table.
// When called by API it responds with person struct without post data merged.
// On frontend call it redirect the client to "/user" page.
func LoginUser(req *http.Request, s sessions.Session, res render.Render, db *r.Session, person Person) {
	person, err := person.Login(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	switch root(req) {
	case "api":
		s.Set("user", person.ID)
		res.JSON(200, person)
		return
	case "user":
		s.Set("user", person.ID)
		res.Redirect("/user", 302)
		return
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

// Login or person.Login is a function which retrieves user according to given .Email field.
// The function then compares the retrieved object's .Digest field with given .Password field.
// If the .Password and .Hash match, the function returns the requested Person struct.
func (person Person) Login(db *r.Session) (Person, error) {
	row, err := r.Table("users").Filter(func(post r.RqlTerm) r.RqlTerm {
		return post.Field("email").Eq(person.Email)
	}).RunRow(db)
	if err != nil || row.IsNil() {
		log.Println(err)
		return person, err
	}
	err = row.Scan(&person)
	if err != nil {
		log.Println(err)
		return person, err
	}
	if CompareHash(person.Digest, person.Password) {
		return person, nil
	}
	return person, errors.New("wrong username or password")
}

// Get or person.Get returns Person object according to given .Id
// with post information merged, but without the .Digest and .Email field.
func (person Person) Get(db *r.Session) (Person, error) {
	row, err := r.Table("users").Get(person.ID).Merge(map[string]interface{}{"posts": r.Table("posts").Filter(func(post r.RqlTerm) r.RqlTerm {
		return post.Field("author").Eq(person.ID)
	}).OrderBy(r.Desc("date")).CoerceTo("ARRAY").Without("author")}).Without("digest", "email").RunRow(db)
	if err != nil {
		log.Println(err)
		return person, err
	}
	if row.IsNil() {
		return person, errors.New("nothing was found")
	}
	err = row.Scan(&person)
	if err != nil {
		log.Println(err)
		return person, err
	}
	return person, err
}

// Session or person.Session returns Person object from client session cookie.
// The returned object has post data merged.
func (person Person) Session(db *r.Session, s sessions.Session) (Person, error) {
	data := s.Get("user")
	id, exists := data.(string)
	if exists {
		var person Person
		person.ID = id
		person, err := person.Get(db)
		if err != nil {
			log.Println(err)
			return person, err
		}
		return person, nil
	}
	return person, nil
}

// Delete or person.Delete deletes the user with given .Id from the database.
func (person Person) Delete(db *r.Session, s sessions.Session) error {
	person, err := person.Session(db, s)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = r.Table("users").Get(person.ID).Delete().RunRow(db)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Insert or person.Insert inserts a new Person struct into the database.
// The function creates .Digest hash from .Password.
func (person Person) Insert(db *r.Session) (Person, error) {
	person.Digest = GenerateHash(person.Password)
	// We dont want to store plaintext password.
	// Options given in Person struct will omit the field
	// from being written to database at all.
	person.Password = ""
	row, err := r.Table("users").Insert(person).RunRow(db)
	if err != nil {
		log.Println(err)
		return person, err
	}
	err = row.Scan(&person)
	if err != nil {
		log.Println(err)
		return person, err
	}
	return person, err
}

// GetAll or person.GetAll fetches all persons with post data merged from the database.
func (person Person) GetAll(db *r.Session) ([]Person, error) {
	var persons []Person
	rows, err := r.Table("users").Without("digest", "email").Run(db)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for rows.Next() {
		err := rows.Scan(&person)
		person, err := person.Get(db)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		persons = append(persons, person)
	}
	return persons, nil
}

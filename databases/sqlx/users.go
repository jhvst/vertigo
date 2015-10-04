package sqlx

import (
	"errors"
	"log"
	"strconv"
	"time"

	. "github.com/9uuso/vertigo/crypto"
	. "github.com/9uuso/vertigo/email"

	"github.com/martini-contrib/sessions"
	"github.com/pborman/uuid"
)

// User struct holds all relevant data for representing user accounts on Vertigo.
// A complete User struct also includes Posts field (type []Post) which includes
// all posts made by the user.
type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name" form:"name"`
	Password string `json:"password,omitempty" form:"password" sql:"-"`
	Recovery string `json:"-"`
	Digest   []byte `json:"-"`
	Email    string `json:"email" form:"email" binding:"required"`
	Posts    []Post `json:"posts"`
	Location string `json:"location" form:"location"`
}

// Login or user.Login is a function which retrieves user according to given .Email field.
// The function then compares the retrieved object's .Digest field with given .Password field.
// If the .Password and .Digest match, the function returns the requested User struct, but with
// the .Password and .Digest omitted.
func (user User) Login() (User, error) {
	password := user.Password
	user, err := user.GetByEmail()
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
func (user User) Update(entry User) (User, error) {
	_, err := db.NamedExec(
		"UPDATE users SET name = :name, digest = :digest, location = :location, recovery = :recovery WHERE id = :id",
		entry)
	if err != nil {
		return entry, err
	}
	return entry, nil
}

// Recover or user.Recover is used to recover User's password according to user.Email
// The function will insert user.Recovery field with generated UUID string and dispatch an email
// to the corresponding user.Email address. It will also add TTL to Recovery field.
func (user User) Recover() error {

	user, err := user.GetByEmail()
	if err != nil {
		return err
	}

	entry := user
	entry.Recovery = uuid.New()
	user, err = user.Update(entry)
	if err != nil {
		return err
	}

	id := strconv.Itoa(int(user.ID))
	err = SendRecoveryEmail(id, user.Name, user.Email, user.Recovery)
	if err != nil {
		return err
	}
	go user.ExpireRecovery(180 * time.Minute)

	return nil
}

func (user User) PasswordReset(entry User) (User, error) {
	digest, err := GenerateHash(entry.Password)
	if err != nil {
		return entry, err
	}
	entry.Digest = digest
	entry.Recovery = ""
	entry.ID = user.ID
	_, err = db.NamedExec("UPDATE users SET digest = :digest, recovery = :recovery WHERE id = :id", entry)
	if err != nil {
		return entry, err
	}
	return entry, nil
}

// ExpireRecovery or user.ExpireRecovery sets a TTL according to t to a recovery hash.
// This function is supposed to be run as goroutine to avoid blocking exection for t.
func (user User) ExpireRecovery(t time.Duration) {
	time.Sleep(t)
	user.Recovery = ""
	_, err := db.NamedExec("UPDATE users SET recovery = :recovery WHERE id = :id", user)
	if err != nil {
		log.Println("expire recovery:", err)
	}
}

// Get or user.Get returns user according to given user.Slug.
// Requires session session as a parameter.
// Returns Ad and error object.
func (user User) Get() (User, error) {
	stmt, err := db.PrepareNamed("SELECT * FROM users WHERE id = :id")
	if err != nil {
		return user, err
	}
	err = stmt.Get(&user, user)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return user, errors.New("not found")
		}
		return user, err
	}
	var posts []Post
	stmt, err = db.PrepareNamed("SELECT * FROM posts WHERE author = :id ORDER BY created")
	if err != nil {
		return user, err
	}
	err = stmt.Select(&posts, user)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			user.Posts = make([]Post, 0)
			return user, nil
		} else {
			return user, err
		}
	}
	user.Posts = posts
	return user, nil
}

// GetByEmail or user.GetByEmail returns User object according to given .Email
// with post information merged.
func (user User) GetByEmail() (User, error) {
	stmt, err := db.PrepareNamed("SELECT * FROM users WHERE email = :email")
	if err != nil {
		return user, err
	}
	err = stmt.Get(&user, user)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return user, errors.New("not found")
		}
		return user, err
	}
	var posts []Post
	stmt, err = db.PrepareNamed("SELECT * FROM posts WHERE author = :id ORDER BY created")
	if err != nil {
		return user, err
	}
	err = stmt.Select(&posts, user)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			user.Posts = make([]Post, 0)
			return user, nil
		} else {
			return user, err
		}
	}
	user.Posts = posts
	return user, nil
}

// Session or user.Session returns user.ID from client session cookie.
// The returned object has post data merged.
func (user User) Session(s sessions.Session) (User, error) {
	data := s.Get("user")
	id, exists := data.(int64)
	if exists {
		var user User
		user.ID = id
		user, err := user.Get()
		if err != nil {
			return user, err
		}
		return user, nil
	}
	return user, errors.New("unauthorized")
}

// Insert or user.Insert inserts a new User struct into the database.
// The function creates .Digest hash from .Password.
func (user User) Insert() (User, error) {
	digest, err := GenerateHash(user.Password)
	if err != nil {
		return user, err
	}
	_, err = time.LoadLocation(user.Location)
	if err != nil {
		return user, errors.New("user location invalid")
	}
	user.Digest = digest
	_, err = db.NamedExec("INSERT INTO users (name, digest, email, location) VALUES (:name, :digest, :email, :location)", user)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: users.email" || err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return user, errors.New("user email exists")
		}
		return user, err
	}
	return user, nil
}

// GetAll or user.GetAll fetches all users with post data merged from the database.
func (user User) GetAll() ([]User, error) {
	var users []User
	rows, err := db.Queryx("SELECT * FROM users")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			users = make([]User, 0)
			return users, nil
		}
		return users, err
	}
	for rows.Next() {
		err := rows.StructScan(&user)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}
	for index, user := range users {
		user, err := user.Get()
		if err != nil {
			return users, err
		}
		if len(user.Posts) == 0 {
			user.Posts = make([]Post, 0)
		}
		users[index] = user
	}
	return users, nil
}

// This file contains about everything related to users aka users. At the top you will find routes
// and at the bottom you can find CRUD options. Some functions in this file are analogous
// to the ones in posts.go.
package gorm

import (
	"errors"
	"log"
	"time"
	"strconv"

	. "github.com/9uuso/vertigo/crypto"
	. "vertigo/email"

	"code.google.com/p/go-uuid/uuid"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/sessions"
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
	query := connection.Gorm.Where(&User{ID: user.ID}).Find(&user).Updates(entry)
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
func (user User) Recover() error {

	user, err := user.GetByEmail()
	if err != nil {
		return err
	}

	var entry User
	entry.Recovery = uuid.New()
	user, err = user.Update(entry)
	if err != nil {
		return err
	}

	id := strconv.Itoa(int(user.ID))
	err = SendRecoveryEmail(id, user.Email, user.Recovery)
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
	entry.Recovery = " " // gorm can't save an empty string
	_, err = user.Update(entry)
	if err != nil {
		return entry, err
	}
	return entry, nil
}

// ExpireRecovery or user.ExpireRecovery sets a TTL according to t to a recovery hash.
// This function is supposed to be run as goroutine to avoid blocking exection for t.
func (user User) ExpireRecovery(t time.Duration) {
	time.Sleep(t)

	var entry User
	entry.Recovery = " "
	_, err := user.Update(entry)
	if err != nil {
		log.Println(err)
	}
	return
}

// Get or user.Get returns User object according to given .ID
// with post information merged.
func (user User) Get() (User, error) {
	var posts []Post
	query := connection.Gorm.Where(&User{ID: user.ID}).First(&user)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			return user, errors.New("not found")
		}
		return user, query.Error
	}
	query = connection.Gorm.Order("created desc").Where(&Post{Author: user.ID}).Find(&posts)
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
func (user User) GetByEmail() (User, error) {
	var posts []Post
	query := connection.Gorm.Where(&User{Email: user.Email}).First(&user)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			return user, errors.New("not found")
		}
		return user, query.Error
	}
	query = connection.Gorm.Where(&Post{Author: user.ID}).Find(&posts)
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

// Delete or user.Delete deletes the user with given ID from the database.
// func (user User) Delete(s sessions.Session) error {
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
func (user User) Insert() (User, error) {
	digest, err := GenerateHash(user.Password)
	if err != nil {
		return user, err
	}
	user.Digest = digest
	user.Posts = make([]Post, 0)
	query := connection.Gorm.Create(&user)
	if connection.Gorm.NewRecord(user) == true {
		return user, errors.New("user email exists")
	}
	if query.Error != nil {
		return user, query.Error
	}
	return user, nil
}

// GetAll or user.GetAll fetches all users with post data merged from the database.
func (user User) GetAll() ([]User, error) {
	var users []User
	query := connection.Gorm.Find(&users)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			users = make([]User, 0)
			return users, nil
		}
		return users, query.Error
	}
	for index, user := range users {
		user, err := user.Get()
		if err != nil {
			return users, err
		}
		users[index] = user
	}
	return users, nil
}

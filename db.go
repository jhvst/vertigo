package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"errors"
	"os"
)


//Returns a martini context to handle middleware calls.
func MongoMiddleware() martini.Handler {
	session, err := mgo.Dial(os.Getenv("DEVDB"))
	if err != nil {
		panic(err)
	}

	return func(c martini.Context) {
		s := session.Clone()
		c.Map(s.DB(os.Getenv("DEVDBNAME")))
	}
}

//Returns a database object for testing purposes, etc. where middleware doesnt reach.
func MongoDB() (db *mgo.Database) {
	session, err := mgo.Dial(os.Getenv("DEVDB"))
	if err != nil {
		panic(err)
	}
	s := session.Clone()
	return s.DB(os.Getenv("DEVDBNAME"))
}

func GetAll(db *mgo.Database) []Person {
	var person []Person
	db.C("users").Find(nil).All(&person)
	return person
}

func GetUser(db *mgo.Database, person Person) Person {
	db.C("users").FindId(person.Id)
	return person
}

func GetUserByID(db *mgo.Database, id string) Person {
	var person Person
	person.Id = bson.ObjectIdHex(id)
	db.C("users").Find(bson.M{"_id": person.Id}).One(&person)
	return person
}

func RemoveUserByID(db *mgo.Database, person *Person) error {
	err := db.C("users").Remove(bson.M{"_id": person.Id})
	if err != nil {
		return err
	}
	return nil
}

func UpdateUserByID(db *mgo.Database, person *Person) (Person, error) {
	err := db.C("users").Update(bson.M{"_id": person.Id}, person)
	if err != nil {
		return *person, err
	}
	return *person, nil
}

func GetUserWithEmail(db *mgo.Database, person *Person) (Person, error) {
	err := db.C("users").Find(bson.M{"email": person.Email}).One(person)
	if err != nil {
		return *person, err
	}
	return *person, nil
}

func GetPostWithTitle(db *mgo.Database, post *Post) (Post, error) {
	err := db.C("posts").Find(bson.M{"title": post.Title}).One(post)
	if err != nil {
		return *post, err
	}
	return *post, nil
}

func RemoveUserBySession(db *mgo.Database, session sessions.Session) error {
	data := session.Get("user")
	email, exists := data.(string)
	if exists {
		var person Person
		person.Email = email
		person, err := GetUserWithEmail(db, &person)
		if err != nil {
			return err
		}
		err = RemoveUserByID(db, &person)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Session could not be retrieved.")
}

func GetPostsFromAuthor(db *mgo.Database, person Person) (posts []Post, err error) {
	err = db.C("posts").Find(bson.M{"author": person.Email}).All(&posts)
	if err != nil {
		return posts, err
	}
	return posts, nil
}

func UpdateUserBySession(db *mgo.Database, session sessions.Session, person Person) error {
	data := session.Get("user")
	email, exists := data.(string)
	if exists && email == person.Email {
		_, err := UpdateUserByID(db, &person)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("There was an error with session. Please log in again and try again.")
}

func GetUserFromSession(db *mgo.Database, session sessions.Session) (Person, error) {
	var person Person
	data := session.Get("user")
	email, exists := data.(string)
	if exists {
		var person Person
		person.Email = email
		person, err := GetUserWithEmail(db, &person)
		if err != nil {
			return person, err
		}
		return person, nil
	}
	return person, errors.New("Session could not be retrieved.")
}

//Part of structure validation — here because it shares the database libaries.
//Checks for collaterating email addresses, which are the 'primary' keys of our structure for now.
func EmailIsUnique(person Person) (unique bool) {
	db := MongoDB()
	count, err := db.C("users").Find(bson.M{"email": person.Email}).Count()
	if err != nil {
		panic(err)
	}
	if count >= 1 {
		return false
	}
	db.Session.Close()
	return true
}

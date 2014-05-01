package main

import (
	r "github.com/dancannon/gorethink"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	//"time"
	"errors"
	"log"
)

func CreateUser(res render.Render, db *r.Session, s sessions.Session, person Person) {
	if !EmailIsUnique(db, person) {
		res.JSON(422, map[string]interface{}{"error": "Email already in use"})
		return
	}
	user, err := person.Insert(db)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	s.Set("user", user.Id)
	res.JSON(200, user)
}

func DeleteUser(res render.Render, db *r.Session, s sessions.Session, person Person) {
	person, err := person.Login(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	err = person.Delete(db, s)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	res.JSON(200, map[string]interface{}{"status": "User successfully deleted"})
}

func ReadUser(params martini.Params, res render.Render, db *r.Session) {
	var person Person
	person.Id = params["id"]
	user, err := person.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	res.JSON(200, user)
}

func ReadUsers(res render.Render, db *r.Session) {
	var person Person
	users, err := person.GetAll(db)
	if err != nil {
		res.JSON(500, err)
		return
	}
	res.JSON(200, users)
}

func EmailIsUnique(s *r.Session, person Person) (unique bool) {
	row, err := r.Table("users").Filter(func(user r.RqlTerm) r.RqlTerm {
		return user.Field("email").Eq(person.Email)
	}).RunRow(s)
	if err != nil || !row.IsNil() {
		return false
	}
	return true
}

func (person Person) Login(s *r.Session) (Person, error) {
	row, err := r.Table("users").Filter(func(post r.RqlTerm) r.RqlTerm {
		return post.Field("email").Eq(person.Email)
	}).RunRow(s)
	if err != nil || row.IsNil() {
		return person, err
	}
	err = row.Scan(&person)
	if err != nil || !CompareHash(person.Digest, person.Password) {
		return person, err
	}
	return person, nil
}

func (person Person) Get(s *r.Session) (Person, error) {
	row, err := r.Table("users").Get(person.Id).Merge(map[string]interface{}{"posts": r.Table("posts").Filter(func(post r.RqlTerm) r.RqlTerm {
		return post.Field("author").Eq(person.Id)
	}).CoerceTo("ARRAY").Without("author")}).Without("digest", "email").RunRow(s)
	if err != nil {
		return person, err
	}
	if row.IsNil() {
		return person, errors.New("Nothing was found.")
	}
	err = row.Scan(&person)
	if err != nil {
		return person, err
	}
	return person, err
}

func (person Person) Session(db *r.Session, s sessions.Session) (Person, error) {
	data := s.Get("user")
	id, exists := data.(string)
	if exists {
		var person Person
		person.Id = id
		person, err := person.Get(db)
		if err != nil {
			return person, err
		}
		return person, nil
	}
	return person, errors.New("Session could not be retrieved.")
}

func (person Person) Delete(db *r.Session, s sessions.Session) error {
	person, err := person.Session(db, s)
	if err != nil {
		return err
	}
	_, err = r.Table("users").Get(person.Id).Delete().RunRow(db)
	if err != nil {
		return err
	}
	return nil
}

func (person Person) Insert(s *r.Session) (Person, error) {
	person.Digest = GenerateHash(person.Password)
	person.Password = ""
	row, err := r.Table("users").Insert(person).RunRow(s)
	if err != nil {
		return person, err
	}
	err = row.Scan(&person)
	if err != nil {
		return person, err
	}
	return person, err
}

func (person Person) GetAll(s *r.Session) ([]Person, error) {
	var persons []Person
	rows, err := r.Table("users").Without("digest", "email").Run(s)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err := rows.Scan(&person)
		person, err := person.Get(s)
		if err != nil {
			return nil, err
		}
		persons = append(persons, person)
	}
	return persons, nil
}

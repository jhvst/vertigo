package main

import (
	r "github.com/dancannon/gorethink"
	//"github.com/go-martini/martini"
	//"time"
	"errors"
	"log"
)

func (person Person) Get(s *r.Session) (Person, error) {
	row, err := r.Table("users").Get(person.Id).EqJoin("author_id", r.Table("posts")).Zip().RunRow(s)
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

func (person Person) GetAll(s *r.Session) ([]Person, error) {
	var persons []Person
	rows, err := r.Table("users").Run(s)
	if err != nil {
		return nil, err
	}
    for rows.Next() {
        err := rows.Scan(&person)
        if err != nil {
            return nil, err
        }
        persons = append(persons, person)
    }
    return persons, nil
}
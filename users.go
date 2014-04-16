package main

import (
	r "github.com/dancannon/gorethink"
	//"github.com/go-martini/martini"
	//"time"
	"errors"
)

func (person Person) Get(s *r.Session) (Person, error) {
	row, err := r.Db("workki").Table("users").Get(person.Id).Merge(map[string]interface{}{"posts":r.Db("workki").Table("posts").Filter(func (post r.RqlTerm) r.RqlTerm {
    	return post.Field("author").Eq(person.Id)
	}).CoerceTo("ARRAY")}).RunRow(s)
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
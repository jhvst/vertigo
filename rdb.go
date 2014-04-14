package main

import (
	r "github.com/dancannon/gorethink"
	"github.com/go-martini/martini"
	"time"
	"errors"
	"os"
)

var session *r.Session

type RethinkDB interface {
	Insert()
	Get()
}

//Insert is alias for
//		r.Table(table).Insert(object).RunWrite(s)
//The function returns yielded structure passed as `object`.
//Table is table name and object is the interface the response will be yielded.
func Insert(s *r.Session, table string, object interface{}) (error) {
	_, err := r.Table(table).Insert(object).RunWrite(s)
	if err != nil {
		return err
	}
	return nil
}

//Get is an alias for
//		r.Table(table).Get(query).RunRow(s)
//Searches table for primary key `query` and returns a single row corresponding to it.
//The query has to match the primary key value of the table for the function to succee.
func Get(s *r.Session, table string, query string, object interface{}) (interface{}, error) {
	row, err := r.Table(table).Get(query).RunRow(s)
	if err != nil {
		return object, err
	}
	if row.IsNil() {
		return object, errors.New("Nothing was found.")
	}
	err = row.Scan(&object)
	if err != nil {
		return object, err
	}
	return object, err
}

func rGetAll(s *r.Session, table string) ([]Person, error) {
	var persons []Person
	rows, err := r.Table(table).Run(s)
	if err != nil {
		return nil, err
	}
    for rows.Next() {
    	var person Person
        err := rows.Scan(&person)
        if err != nil {
            return nil, err
        }
        persons = append(persons, person)
    }
    return persons, nil
}

//Middleware function hooks the RethinkDB to be accessible for Martini routes.
//By deafault the middleware spawns a session pool of 10 connections.
//Typical connection options on development environment would be
//		Address: "localhost:28015"
//		Database: "test"
func Middleware() martini.Handler {
	session, err := r.Connect(r.ConnectOpts{
	    Address:  os.Getenv("rDB"),
	    Database: os.Getenv("rNAME"),
	    MaxIdle: 10,
	    IdleTimeout: time.Second * 10,
	})

	if err != nil {
	    panic(err)
	}

	return func(c martini.Context) {
		c.Map(session)
	}
}
package main

import (
    r "github.com/dancannon/gorethink"
    "log"
)

func main() {
	var session *r.Session

	session, err := r.Connect(r.ConnectOpts{
	    Address:  "localhost:28015",
	})

	r.DbCreate("vertigo").RunRow(session)

	if err != nil {
	    log.Fatalln(err.Error())
	}

	_, err = r.Db("vertigo").TableCreate("users").RunWrite(session)
	if err != nil {
	    log.Fatalf("Error creating table: %s", err)
	}

	_, err = r.Db("vertigo").TableCreate("posts").RunWrite(session)
	if err != nil {
	    log.Fatalf("Error creating table: %s", err)
	}

	log.Println("Database successfully created and tables initated.")
}
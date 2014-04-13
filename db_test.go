package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"testing"
)

func TestSpec(t *testing.T) {
	var person Person

	person.Email = "juuso@mail.com"

	// Only pass t into top-level Convey calls
	Convey("Given person with starting values", t, func() {
		person, err := GetUserWithEmail(MongoDB(), &person)
		if err != nil {
			log.Println(err)
		}
		So(person.Email, ShouldEqual, "juuso@mail.com")
	})

}

package main

import (
    "testing"
    "log"
    "os"
)

func Test_Martini_RunOnAddr(t *testing.T) {
	// just test that Run doesn't bomb
	if os.Getenv("WERCKER_RETHINKDB_HOST") != "" {
		os.Setenv("RDB_HOST", os.Getenv("WERCKER_RETHINKDB_HOST"))
	}
	if os.Getenv("WERCKER_RETHINKDB_PORT") != "" {
		os.Setenv("RDB_PORT", os.Getenv("WERCKER_RETHINKDB_PORT"))
	}
	go main()
}

func Test_Settings_Creation(t *testing.T) {
	settings := VertigoSettings()
	if settings.Firstrun == true {
		log.Println("no errors")
	}
}
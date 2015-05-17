// This file contains bunch of miscful helper functions.
// The functions here are either too rare to be assiociated to some known file
// or are met more or less everywhere across the code.
package gorm

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Connection struct {
	SQL  *sql.DB
	Gorm gorm.DB
}

var driver = flag.String("driver", "sqlite3", "name of the database driver used, by default sqlite3")
var dbsource = flag.String("dbsource", "./vertigo.db", "connection string or path to database file")
var connection Connection

func init() {

	if os.Getenv("DATABASE_URL") != "" {
		flag.Set("driver", "postgres")
		flag.Set("dbsource", os.Getenv("DATABASE_URL"))
		log.Println("Using PostgreSQL")
	} else {
		log.Println("Using SQLite3")
	}

	db, err := gorm.Open(*driver, *dbsource)

	if err != nil {
		panic(err)
	}

	db.LogMode(false)

	// Here database and tables are created in case they do not exist yet.
	// If database or tables do exist, nothing will happen to the original ones.
	db.CreateTable(&User{})
	db.CreateTable(&Post{})
	db.AutoMigrate(&User{}, &Post{})

	connection.SQL = db.DB()
	connection.Gorm = db
}

// Open is generic function to open database for gorm. By default, Open is only
// used in tests.
func Open() gorm.DB {
	db, err := gorm.Open(*driver, *dbsource)
	if err != nil {
		panic(err)
	}
	connection.SQL = db.DB()
	connection.Gorm = db
	return db
}

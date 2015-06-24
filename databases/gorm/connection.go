// connection.go stores the Gorm specific database connection settings,
// including the init() function
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

// Connection structure stores the raw SQL pointer of the database connection
// for simpler and easier use on User and Post methods.
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

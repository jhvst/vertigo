package sqlx

import (
	"flag"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

var sqlite3 = `
CREATE TABLE users (
    id integer NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL,
    recovery char(36) NOT NULL DEFAULT "",
    digest blob NOT NULL,
    email varchar(255) NOT NULL UNIQUE,
    location varchar(255) NOT NULL DEFAULT "UTC"
);

CREATE TABLE posts (
    id integer NOT NULL PRIMARY KEY,
    title varchar(255) NOT NULL,
    content text NOT NULL,
    markdown text NOT NULL,
    slug varchar(255) NOT NULL,
    author integer NOT NULL,
    excerpt varchar(255) NOT NULL,
    viewcount integer unsigned NOT NULL DEFAULT 0,
    published bool NOT NULL DEFAULT false,
    created integer unsigned NOT NULL,
    updated integer unsigned NOT NULL,
    timeoffset integer NOT NULL DEFAULT 0
);`

var postgres = `
CREATE TABLE "users" (
    "id" serial NOT NULL PRIMARY KEY,
    "name" varchar(255) NOT NULL,
    "recovery" char(36) NOT NULL DEFAULT '',
    "digest" bytea NOT NULL,
    "email" varchar(255) NOT NULL UNIQUE,
    "location" varchar(255) NOT NULL DEFAULT 'UTC'
);

CREATE TABLE "posts" (
    "id" serial NOT NULL PRIMARY KEY,
    "title" varchar(255) NOT NULL,
    "content" text NOT NULL,
    "markdown" text NOT NULL,
    "slug" varchar(255) NOT NULL,
    "author" integer NOT NULL,
    "excerpt" varchar(255) NOT NULL,
    "viewcount" integer NOT NULL DEFAULT '0',
    "published" bool NOT NULL DEFAULT false,
    "created" integer NOT NULL,
    "updated" integer NOT NULL,
    "timeoffset" integer NOT NULL DEFAULT '0'
);`

// var mysql = `
// CREATE DATABASE vertigo;
// USE vertigo;

// CREATE TABLE IF NOT EXISTS users (
//     id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
//     name varchar(255) NOT NULL,
//     recovery char(36) NOT NULL DEFAULT "",
//     digest blob NOT NULL,
//     email varchar(255) NOT NULL UNIQUE,
//     location varchar(255) NOT NULL DEFAULT "UTC"
// );

// CREATE TABLE IF NOT EXISTS posts (
//     id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
//     title varchar(255) NOT NULL,
//     content text NOT NULL,
//     markdown text NOT NULL,
//     slug varchar(255) NOT NULL,
//     author integer NOT NULL,
//     excerpt varchar(255) NOT NULL,
//     viewcount integer unsigned NOT NULL DEFAULT 0,
//     published bool NOT NULL DEFAULT false,
//     created integer unsigned NOT NULL,
//     updated integer unsigned NOT NULL,
//     timeoffset integer NOT NULL DEFAULT 0
// );`

func Drop() {
	db.MustExec("DROP TABLE users")
	db.MustExec("DROP TABLE posts")
	os.Remove("settings.json")
	os.Remove("vertigo.db")
}

var Driver = flag.String("driver", "sqlite3", "Database driver to use (sqlite3, mysql, postgres)")
var Source = flag.String("source", "vertigo.db", "Database data source")

func init() {

	flag.Parse()

    if os.Getenv("DATABASE_URL") != "" {
        *Driver = "postgres"
        *Source = os.Getenv("DATABASE_URL")
    }    

	conn, err := sqlx.Connect(*Driver, *Source)
	if err != nil {
		log.Fatal("sqlx connect:", err)
	}

	var schema string
	switch *Driver {
	case "sqlite3":
		schema = sqlite3
	// case "mysql":
	//     schema = mysql
	case "postgres":
		schema = postgres
	}

	conn.Exec(schema)

	log.Println("sqlx: using", *Driver)

	db = conn
}

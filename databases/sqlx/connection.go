package sqlx

import (
	"flag"
	"log"
	"net/url"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB
var Settings *Vertigo

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
    slug varchar(255) NOT NULL UNIQUE,
    author integer NOT NULL,
    excerpt varchar(255) NOT NULL,
    viewcount integer unsigned NOT NULL DEFAULT 0,
    published bool NOT NULL DEFAULT false,
    created integer unsigned NOT NULL,
    updated integer unsigned NOT NULL,
    timeoffset integer NOT NULL DEFAULT 0
);

CREATE TABLE settings (
    id integer NOT NULL PRIMARY KEY DEFAULT 1,
    name varchar(255) NOT NULL,
    hostname varchar(255) NOT NULL,
    firstrun bool NOT NULL DEFAULT true,
    cookiehash string NOT NULL,
    allowregistrations bool NOT NULL DEFAULT true,
    description varchar(255) NOT NULL,
    mailerlogin varchar(255),
    mailerport integer unsigned NOT NULL DEFAULT 587,
    mailerpassword varchar(255),
    mailerhostname varchar(255)
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
    "slug" varchar(255) NOT NULL UNIQUE,
    "author" integer NOT NULL,
    "excerpt" varchar(255) NOT NULL,
    "viewcount" integer NOT NULL DEFAULT '0',
    "published" bool NOT NULL DEFAULT false,
    "created" integer NOT NULL,
    "updated" integer NOT NULL,
    "timeoffset" integer NOT NULL DEFAULT '0'
);

CREATE TABLE "settings" (
    "id" serial NOT NULL PRIMARY KEY,
    "name" varchar(255) NOT NULL,
    "hostname" varchar(255) NOT NULL,
    "firstrun" bool NOT NULL DEFAULT true,
    "cookiehash" bytea NOT NULL,
    "allowregistrations" bool NOT NULL DEFAULT true,
    "description" varchar(255) NOT NULL,
    "mailerlogin" varchar(255),
    "mailerport" integer NOT NULL DEFAULT 587,
    "mailerpassword" varchar(255),
    "mailerhostname" varchar(255)
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
	db.MustExec("DROP TABLE settings")
	os.Remove("vertigo.db")
}

var Driver = flag.String("driver", "sqlite3", "Database driver to use (sqlite3, mysql, postgres)")
var Source = flag.String("source", "vertigo.db", "Database data source")

func connect(driver, source string) {
	conn, err := sqlx.Connect(driver, source)
	if err != nil {
		log.Fatal("sqlx connect:", err)
	}

	var schema string
	switch driver {
	case "sqlite3":
		schema = sqlite3
	/*case "mysql":
	schema = mysql*/
	case "postgres":
		schema = postgres
	}

	conn.MustExec(schema)

	log.Println("sqlx: using", driver)

	db = conn

	Settings = VertigoSettings()
}

func init() {

	flag.Parse()

	if os.Getenv("DATABASE_URL") != "" {
		u, err := url.Parse(os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatal("database url parameter could not be parsed")
		}
		if u.Scheme != "postgres" /* && u.Scheme != "mysql" */ {
			log.Fatal("unsupported database type")
		}
		connect(u.Scheme, os.Getenv("DATABASE_URL"))
		return
	}

	connect(*Driver, *Source)
}

package sqlx

import (
	"log"
	"os"
    "flag"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"
)

var db *sqlx.DB

var sqlite3 = `
CREATE TABLE user (
    id integer NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL,
    recovery char(36) NOT NULL DEFAULT "",
    digest blob NOT NULL,
    email varchar(255) NOT NULL UNIQUE,
    location varchar(255) NOT NULL DEFAULT "UTC"
);

CREATE TABLE post (
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
CREATE TABLE user (
    id serial NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL,
    recovery char(36) NOT NULL DEFAULT "",
    digest blob NOT NULL,
    email varchar(255) NOT NULL UNIQUE,
    location varchar(255) NOT NULL DEFAULT "UTC"
);

CREATE TABLE post (
    id serial NOT NULL PRIMARY KEY,
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

var mysql = `
CREATE DATABASE vertigo;
USE vertigo;

CREATE TABLE IF NOT EXISTS user (
    id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name varchar(255) NOT NULL,
    recovery char(36) NOT NULL DEFAULT "",
    digest blob NOT NULL,
    email varchar(255) NOT NULL UNIQUE,
    location varchar(255) NOT NULL DEFAULT "UTC"
);

CREATE TABLE IF NOT EXISTS post (
    id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
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

func Drop() {
	db.MustExec("DROP TABLE user")
	db.MustExec("DROP TABLE post")
	os.Remove("test.db")
}

func init() {

    driver := flag.String("driver", "sqlite3", "Database driver to use (sqlite3, mysql, postgres)")
    source := flag.String("source", "vertigo.db", "Database data source")

    flag.Parse()   

	conn, err := sqlx.Connect(*driver, *source)
	if err != nil {
		log.Fatal("sqlx connect:", err)
	}

    var schema string
    switch (*driver) {
    case "sqlite3":
        schema = sqlite3
    case "mysql":
        schema = mysql
    case "postgres":
        schema = postgres
    }

	conn.Exec(schema)

    log.Println("sqlx: using", *driver)

	db = conn
}

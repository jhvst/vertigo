package sqlx

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

var schema = `
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
    timeoffset integer NOT NULL DEFAULT 0,
    FOREIGN KEY (author) REFERENCES user (id)
);`

func Drop() {
	db.MustExec("DROP TABLE user")
	db.MustExec("DROP TABLE post")
	os.Remove("test.db")
}

func init() {
	conn, err := sqlx.Connect("sqlite3", "vertigo.db")
	if err != nil {
		log.Fatal("sqlx connect:", err)
	}
	conn.MustExec(schema)

	db = conn
}

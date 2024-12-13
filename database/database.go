package database

import (
	"database/sql"
	"fmt"

	_ "github.com/glebarez/go-sqlite"
)

func SqliteDB() *sql.DB {
	db, err := sql.Open("sqlite", "./goquestdb.db")
	if err != nil {
		fmt.Println("Error opening sqlite")
	}
	fmt.Println("Conected to sql sucessifully")
	return db

}

func Migrations(db *sql.DB) {

	fmt.Println("preparing to create table")
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS requests (id INTEGER PRIMARY KEY, name VARCHAR(64), method VARCHAR(64), route VARCHAR(64), params VARCHAR(64) NULL, headers VARCHAR(64) NULL)")
	if err != nil {
		fmt.Println("Error preparing the table creation statement:", err)
		return
	}
	defer statement.Close()

	_, err = statement.Exec()
	if err != nil {
		fmt.Println("Error executing the statement:", err)
		return
	}
	fmt.Println("Successfully created table requests!")

}


package main

import "fmt"

/*
CREATE TABLE account
	id INTEGER NOT NULL PRIMARY KEY,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	name TEXT,
	email TEXT,
	password TEXT
	is_admin INTEGER
*/

func CreateDatabase() {
	fmt.Print("Creating database...")
}

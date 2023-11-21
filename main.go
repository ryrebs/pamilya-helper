package main

import (
	"log"
	"os"

	"pamilyahelper/webapp/server"
	"pamilyahelper/webapp/server/db"
)

const helpMsg = `
PamilyaHelper db tool.

Example: manage initdb

Command arguments:

	- runserver: Run the server.

	- initdb: Creates and initializes the database.
	
	- loadfixture: Loads fixtures.

	- createadmin: Create admin with default username and password.

`

func main() {
	args := os.Args

	if len(args) < 2 {
		log.Println("See arguments with: `manage help`")
		return
	}

	switch arg := args[1]; arg {
	case "initdb":
		db.InitDB()
	case "loadfixtures":
		db.LoadFixtures()
	case "createadmin":
		db.CreateDefaultAdmin()
	case "serve":
		server.Serve()
	case "help":
		log.Print(helpMsg)
	default:
		log.Println("Command not found. See help with: `manage help`")
	}

}

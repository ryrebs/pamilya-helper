package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
)

const initSqlStmt = `
CREATE TABLE IF NOT EXISTS account (
	id INTEGER NOT NULL PRIMARY KEY,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	name TEXT,
	email TEXT,
	password TEXT
	is_admin INTEGER
)
`

const DefaultPamilyaHelperDBName = "pamilyahelper.db"

type CustomDBContext struct {
	echo.Context
	db *sql.DB
}

func (c *CustomDBContext) Db() *sql.DB {
	return c.db
}

func AddDBContextMiddleware(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomDBContext{c, db}
			return next(cc)
		}
	}
}

func GetDBConn(dbName string) *sql.DB {
	db, err := sql.Open("sqlite3", dbName)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return db
}

func InitDB() *sql.DB {
	os.Remove(DefaultPamilyaHelperDBName)
	db := GetDBConn(DefaultPamilyaHelperDBName)

	if db != nil {
		_, err := db.Exec(initSqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, initSqlStmt)
			return nil
		}
	}

	return db
}

func LoadFixtures() {

}

func CreateDefaultAdmin() {

}

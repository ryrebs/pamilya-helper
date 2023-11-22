package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const initSqlStmt = `
CREATE TABLE IF NOT EXISTS account (
	id INTEGER NOT NULL PRIMARY KEY,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	name TEXT,
	email TEXT UNIQUE,
	password TEXT,
	birthdate DATE,
	address TEXT,
	id_type TEXT,
	is_verified INTEGER,
	is_admin INTEGER
);
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
		} else {
			log.Printf("Successfully initialized database with queries:\n%s", initSqlStmt)
		}
	}

	return db
}

func CreateUserPassword(password string) string {
	pwd := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)

	if err != nil {
		panic(err)
	}

	return string(hashedPassword)
}

func IsPWDvalid(password, hashedStr string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedStr), []byte(password)); err == nil {
		return true
	}
	return false
}

func createAdmin(db *sql.DB) {
	fixtureAdminStmt := `
	INSERT INTO account(name, email, password, birthdate, is_verified, is_admin)
	 	VALUES('admin', 'admin@pamilyahelper.com', ?, '1992-01-01', 1, 1)
	`

	if db != nil {
		stmt, err := db.Prepare(fixtureAdminStmt)
		if err == nil {
			_, err = stmt.Exec(CreateUserPassword("admin"))
			if err == nil {
				log.Println("Created 'admin' user with 'admin' as password...")
				return
			}
			log.Printf("%q: %s\n", err, fixtureAdminStmt)
		}
		log.Printf("%q: %s\n", err, fixtureAdminStmt)
	}
}

func LoadFixtures() {
	db := GetDBConn(DefaultPamilyaHelperDBName)
	createAdmin(db)

}

func CreateDefaultAdmin() {

}

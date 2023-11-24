package db

import (
	"database/sql"
	"errors"
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
	is_verified INTEGER DEFAULT 0,
	is_verification_pending INTEGER DEFAULT 0,
	is_admin INTEGER DEFAULT 0
);
CREATE TABLE IF NOT EXISTS upload (
	id INTEGER NOT NULL PRIMARY KEY,
	image TEXT,
	account_id INTEGER
	detail TEXT
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

func InsertUser(name, email, password string, is_verified, is_admin bool, db *sql.DB) error {
	fixtureAdminStmt := `
		INSERT INTO account(name, email, password, is_verified, is_admin)
		VALUES(?, ?, ?, ?, ?)
	`
	if db != nil {
		stmt, err := db.Prepare(fixtureAdminStmt)
		if err == nil {
			_, err = stmt.Exec(name, email, CreateUserPassword(password), is_verified, is_admin)
			if err == nil {
				return nil
			}
			log.Printf("%q: %s\n", err, fixtureAdminStmt)
		}
		log.Printf("%q: %s\n", err, fixtureAdminStmt)
		return err
	}
	return errors.New("no database connection found")
}

func createAdmin(db *sql.DB) error {
	err := InsertUser("admin", "admin@admin.com", "admin1234", true, true, db)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func LoadFixtures() {
	db := GetDBConn(DefaultPamilyaHelperDBName)
	err := createAdmin(db)
	if err == nil {
		log.Println("Created initial 'admin' user with 'admin1234' as password...")
	}

}

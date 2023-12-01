package db

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"mime/multipart"
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
	birthdate TEXT,
	address TEXT,
	is_verified INTEGER DEFAULT 0,
	is_verification_pending INTEGER DEFAULT 0,
	is_admin INTEGER DEFAULT 0
);
CREATE TABLE IF NOT EXISTS upload (
	id INTEGER NOT NULL PRIMARY KEY,
	image TEXT,
	account_id INTEGER,
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
		defer stmt.Close()
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
		log.Println("Created initial administrator: 'admin' with 'admin1234' as password...")
	}
	err = InsertUser("aubrey", "aubrey@pmh.com", "aubrey1234", false, false, db)
	if err == nil {
		log.Println("Created initial user: 'aubrey' with 'aubrey1234' as password...")
	}
}

func RemoveUser(email string, db *sql.DB) error {
	removeUserStmt := `
		DELETE FROM account WHERE email = ?
	`
	if db != nil {
		stmt, err := db.Prepare(removeUserStmt)
		if err == nil {
			_, err = stmt.Exec(email)
			if err == nil {
				return nil
			}
			log.Printf("%q: %s\n", err, removeUserStmt)
		}
		defer stmt.Close()
		log.Printf("%q: %s\n", err, removeUserStmt)
		return err
	}
	return errors.New("no database connection found")

}

// Update user details
func UpdateUser(email string, details EditableUserFields, db *sql.DB) error {
	if db != nil {
		stmt, err := db.Prepare(`UPDATE account SET name=?,birthdate=?,
									address=?,
									is_verification_pending=1
								 WHERE email=?`)
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = stmt.Exec(details.Name, details.Birthdate, details.Address, email)
		defer stmt.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
	return errors.New("no db found")
}

// Update or insert government id.
func InsertGovID(accountId int, fileName string, db *sql.DB) error {
	if db != nil {
		// Check for existing gov_id
		stmt, err := db.Prepare(`SELECT account_id from upload where account_id=? and detail='gov_id' LIMIT 1`)
		var exists = 0
		if err != nil {
			log.Println(err)
			return err
		}
		err = stmt.QueryRow(accountId).Scan(&exists)
		if err != nil {
			// No record found.
			log.Println(err)
		}
		defer stmt.Close()

		insertStmt := `INSERT INTO upload(image, account_id, detail) VALUES(?, ?, 'gov_id')`
		updatestmt := `UPDATE upload SET image=? WHERE account_id=? AND detail='gov_id'`
		mainStmt := ``

		// Account has existing gov id.
		if exists > 0 {
			mainStmt = updatestmt
		} else {
			mainStmt = insertStmt
		}
		stmt, err = db.Prepare(mainStmt)
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = stmt.Exec(fileName, accountId)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil

	}
	return errors.New("no db found")
}

// Get user's government id
func GetUserGovId(accountId int, db *sql.DB) string {
	stmt, err := db.Prepare(`SELECT image from upload where account_id=?`)
	fileName := ""
	if err != nil {
		log.Println(err)
	}
	err = stmt.QueryRow(accountId).Scan(&fileName)
	if err != nil {
		// No record found.
		log.Println(err)
	}
	defer stmt.Close()
	return fileName
}

// Create file uploads
func CreateFile(file *multipart.FileHeader, filename string, accountId int) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

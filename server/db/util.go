package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

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
		log.Println(err)
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

func InsertUser(user User, db *sql.DB) error {
	fixtureAdminStmt := `
		INSERT INTO account(
			name, email, password,
			is_verified, is_admin, is_verification_pending,
			detail, contact, birthdate, address)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	if db != nil {
		stmt, err := db.Prepare(fixtureAdminStmt)

		if err != nil {
			log.Printf("%q: %s\n", err, fixtureAdminStmt)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(user.Name, user.Email,
			CreateUserPassword(user.Password), user.IsVerified,
			user.IsAdmin, user.IsVerificationPending,
			user.Detail, user.Contact, user.Birthdate, user.Address)
		if err != nil {
			log.Printf("%q: %s\n", err, fixtureAdminStmt)
			return err
		}
		return nil
	}
	return errors.New("no database connection found")
}

func InsertJob(dateline, title, descp, respb, skills, loc, pf, pt, employer_type string, emp_id int, db *sql.DB) error {
	fixtureJobStmt := `
		INSERT INTO job(dateline, title, description, responsibility, skills, location, price_from, price_to, employment_type, employer_id)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	if db != nil {
		stmt, err := db.Prepare(fixtureJobStmt)

		if err != nil {
			log.Printf("%q: %s\n", err, fixtureJobStmt)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(dateline, title, descp, respb, skills, loc, pf, pt, employer_type, emp_id)
		if err != nil {
			log.Printf("%q: %s\n", err, fixtureJobStmt)
			return err
		}
		return nil
	}
	return errors.New("no database connection found")
}

func loadJobFixture(conn *sql.DB) error {
	var jobs struct {
		Jobs []Job `json:"jobs"`
	}
	jsonFile, err := os.Open("fixtures/job.json")
	if err != nil {
		log.Println(err)
		return err
	}
	defer jsonFile.Close()
	content, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Println(err)
		return err
	}
	err = json.Unmarshal(content, &jobs)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, job := range jobs.Jobs {
		err := InsertJob(
			job.DateLine, job.Title,
			job.Description, job.Responsibility, job.Skills,
			job.Location, job.PriceFrom, job.PriceTo,
			job.EmployementType, job.EmployerId, conn)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func loadUserFixture(conn *sql.DB) error {
	var users struct {
		Users []User `json:"users"`
	}
	jsonFile, err := os.Open("fixtures/user.json")
	if err != nil {
		log.Println(err)
		return err
	}
	defer jsonFile.Close()
	content, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Println(err)
		return err
	}
	err = json.Unmarshal(content, &users)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, user := range users.Users {
		err := InsertUser(
			User{
				UserDetail: UserDetail{
					Name:                  user.Name,
					Email:                 user.Email,
					Birthdate:             user.Birthdate,
					Address:               user.Address,
					IsVerified:            user.IsVerified,
					IsAdmin:               user.IsAdmin,
					IsVerificationPending: user.IsVerificationPending,
					Detail:                user.Detail,
					Contact:               user.Contact,
				},
				Password: user.Password,
			},

			conn)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func LoadFixtures() {
	conn := GetDBConn(DefaultPamilyaHelperDBName)
	err := loadUserFixture(conn)
	if err == nil {
		log.Println("Created initial users...")
	}
	err = loadJobFixture(conn)
	if err == nil {
		log.Println("Created initial jobs...")
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

// Get user's government id file name
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

func FindUserFromDb(email string, db *sql.DB) User {
	stmt, err := db.Prepare(`SELECT
								name, email,
								password, birthdate,
								address, is_verified,
								is_admin, id,
								is_verification_pending
							FROM account WHERE email=?`)
	if err != nil {
		log.Println(err)
		return User{}
	}
	defer stmt.Close()
	var user User
	err = stmt.QueryRow(email).Scan(
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Birthdate,
		&user.Address,
		&user.IsVerified,
		&user.IsAdmin,
		&user.AccountId,
		&user.IsVerificationPending)

	if err != nil {
		log.Println(err)
		return User{}
	}
	return user
}

func GetAccountsForVerificationFromDb(limit, offset string, db *sql.DB) ([]UserVerification, error) {
	query := `
		SELECT email, name, birthdate,
			address, image
		FROM account AS ac LEFT JOIN upload AS t ON t.account_id = ac.id
		WHERE t.detail = 'gov_id' AND ac.is_verification_pending = 1
		LIMIT ? OFFSET ?`

	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit, offset)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var users []UserVerification
	for rows.Next() {
		var user UserVerification
		err = rows.Scan(&user.Email, &user.Name, &user.Birthdate, &user.Address, &user.GovId)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return users, nil
}

func UpdateUserVerification(email string, verify int, conn *sql.DB) error {
	if conn != nil {
		updateStmt := `UPDATE account 
						SET is_verified=?,is_verification_pending=0
				   WHERE email=?`
		tx, err := conn.Begin()
		if err != nil {
			log.Println(err)
		}
		stmt, err := tx.Prepare(updateStmt)
		if err != nil {
			log.Println(err)
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(verify, email) // Execute update
		if err != nil {
			log.Println(err)
			return err
		}
		err = tx.Commit()
		if err != nil {
			log.Println(err)
			return err
		}

		return nil
	}
	return errors.New("no db found")
}

// Return jobs that are not posted by the current user and did not submit any application.
//
//	Use customJobQuery if you want to modify the default conditions.
//
//	Should return all the columns of the job in the table.
func GetJobsFromDB(customJobQuery, limit, offset string, accountID int, conn *sql.DB) ([]Job, error) {
	query := `
		SELECT * FROM job WHERE employer_id != ? AND id NOT IN
			(SELECT job_id FROM job_application WHERE employee_id == ?)
		LIMIT ? OFFSET ?
	`
	var err error
	if customJobQuery != "" {
		query = customJobQuery
	}
	stmt, err := conn.Prepare(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	var rows *sql.Rows
	if customJobQuery != "" {
		rows, err = stmt.Query(accountID, limit, offset)

	} else {
		rows, err = stmt.Query(accountID, accountID, limit, offset)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		err = rows.Scan(&job.ID, &job.EmployerId,
			&job.Title, &job.Description, &job.Responsibility,
			&job.Skills, &job.Location, &job.PriceFrom,
			&job.PriceTo, &job.EmployementType, &job.DateLine)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		jobs = append(jobs, job)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return jobs, nil
}

func GetJobFromDB(jobID int, conn *sql.DB) (interface{}, error) {
	query := `SELECT jb.id, employer_id,
					 title, description,
					responsibility, skills,
					location, price_from, price_to,
					employment_type, dateline,
					emp.name, emp.email, emp.contact, emp.detail
			FROM job as jb INNER JOIN account as emp ON jb.employer_id = emp.id
			WHERE jb.id = ?`
	stmt, err := conn.Prepare(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	var job Job
	err = stmt.QueryRow(jobID).Scan(&job.ID, &job.EmployerId,
		&job.Title, &job.Description, &job.Responsibility,
		&job.Skills, &job.Location, &job.PriceFrom,
		&job.PriceTo, &job.EmployementType, &job.DateLine,
		&job.EmployerName, &job.EmployerEmail,
		&job.EmployerContact, &job.EmployerDetail)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return job, nil
}

func InsertJobApplication(jobID, employeeID int, conn *sql.DB) error {
	fixtureJobAppStmt := `INSERT INTO job_application(job_id, employee_id) VALUES(?, ?)`
	if conn != nil {
		stmt, err := conn.Prepare(fixtureJobAppStmt)

		if err != nil {
			log.Printf("%q: %s\n", err, fixtureJobAppStmt)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(jobID, employeeID)
		if err != nil {
			log.Printf("%q: %s\n", err, fixtureJobAppStmt)
			return err
		}
		return nil
	}
	return errors.New("no database connection found")

}

func GetAppliedJobsFromDB(limit, offset string, accountID int, conn *sql.DB) ([]Application, error) {
	query := `SELECT * FROM job jb
		INNER JOIN job_application ja on jb.id = ja.id
		WHERE jb.employer_id != ? AND ja.employee_id == ?
		LIMIT ? OFFSET ?
	`
	stmt, err := conn.Prepare(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(accountID, accountID, limit, offset)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var jobs []Application
	for rows.Next() {
		var job Application
		err = rows.Scan(&job.ID, &job.EmployerId,
			&job.Title, &job.Description, &job.Responsibility,
			&job.Skills, &job.Location, &job.PriceFrom,
			&job.PriceTo, &job.EmployementType, &job.DateLine,
			&job.ApplicationId, &job.Timestamp,
			&job.EmployeeId, &job.JobID, &job.Status)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		jobs = append(jobs, job)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return jobs, nil
}

func IsJobOwned(jobID, accoundID int, conn *sql.DB) (bool, error) {
	query := `SELECT id FROM job WHERE id = ? AND employer_id = ?`
	stmt, err := conn.Prepare(query)
	if err != nil {
		log.Println(err)
	}
	defer stmt.Close()
	var job struct {
		ID int
	}
	err = stmt.QueryRow(jobID, accoundID).Scan(&job.ID)
	if err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "no rows") {
			// When no row is found. Job is not owned by the user.
			return false, nil
		}
	}
	return true, nil
}

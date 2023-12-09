package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"time"

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
			detail, contact, birthdate, address, title, profile_image, skills)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
			user.Detail, user.Contact, user.Birthdate,
			user.Address, user.Title, user.ProfileImage, user.Skills)
		if err != nil {
			log.Printf("%q: %s\n", err, fixtureAdminStmt)
			return err
		}
		return nil
	}
	return errors.New("no database connection found")
}

func InsertJob(dateline, title, descp, respb, skills, loc, pf, pt, employer_type string, emp_id int, db *sql.DB) error {
	insertJobStmt := `
		INSERT INTO job(dateline, title, description, responsibility, skills, location, price_from, price_to, employment_type, employer_id)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	if db != nil {
		stmt, err := db.Prepare(insertJobStmt)

		if err != nil {
			log.Printf("%q: %s\n", err, insertJobStmt)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(dateline, title, descp, respb, skills, loc, pf, pt, employer_type, emp_id)
		if err != nil {
			log.Printf("%q: %s\n", err, insertJobStmt)
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
			return err
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
					ProfileImage:          user.ProfileImage,
					Title:                 user.Title,
					Skills:                user.Skills,
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
func UpdateUser(govIDFileName, email string, details EditableUserFields, db *sql.DB) error {
	if db != nil {
		stmt, err := db.Prepare(`UPDATE account SET name=?,birthdate=?,
									address=?,
									gov_id_image=?,
									is_verification_pending=1
								 WHERE email=?`)
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = stmt.Exec(details.Name, details.Birthdate, details.Address, govIDFileName, email)
		defer stmt.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
	return errors.New("no db found")
}

// Update user profile image
func UpdateProfileImage(userID int, fileName string, db *sql.DB) error {
	if db != nil {
		stmt, err := db.Prepare(`UPDATE account SET profile_image = ? WHERE id=?`)
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = stmt.Exec(fileName, userID)
		defer stmt.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
	return errors.New("no db found")
}

// Update user ITR
func UpdateProfileITR(userID int, fileName string, db *sql.DB) error {
	if db != nil {
		stmt, err := db.Prepare(`UPDATE account SET income_tax_return_file = ? WHERE id=?`)
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = stmt.Exec(fileName, userID)
		defer stmt.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
	return errors.New("no db found")
}

// Returns user detail with type UserDetail if found
func FindUserByIDFromDB(ID int, conn *sql.DB) *UserDetail {
	stmt, err := conn.Prepare(`SELECT
								name, email,
								birthdate, address,
								is_verified,
								is_admin, id,
								is_verification_pending,
								gov_id_image, profile_image,
								detail, title,
								skills,	income_tax_return_file
							FROM account WHERE id=?`)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer stmt.Close()
	var user UserDetail
	err = stmt.QueryRow(ID).Scan(
		&user.Name,
		&user.Email,
		&user.Birthdate,
		&user.Address,
		&user.IsVerified,
		&user.IsAdmin,
		&user.AccountId,
		&user.IsVerificationPending,
		&user.GovIDImage,
		&user.ProfileImage,
		&user.Detail,
		&user.Title,
		&user.Skills,
		&user.IncomeTaxReturnFile)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &user
}

func FindUserFromDb(email string, db *sql.DB) User {
	stmt, err := db.Prepare(`SELECT
								name, email,
								password, birthdate,
								address, is_verified,
								is_admin, id,
								is_verification_pending,
								gov_id_image, profile_image,
								detail, title,
								skills,	income_tax_return_file
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
		&user.IsVerificationPending,
		&user.GovIDImage,
		&user.ProfileImage,
		&user.Detail,
		&user.Title,
		&user.Skills,
		&user.IncomeTaxReturnFile)
	if err != nil {
		log.Println(err)
		return User{}
	}
	return user
}

func GetAccountsForVerificationFromDb(limit, offset string, db *sql.DB) ([]UserVerification, error) {
	query := `
		SELECT email, name, birthdate, address, gov_id_image
		FROM account WHERE is_verification_pending = 1
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

// Get jobs in the database regardless of the owner
func GetJobsAllJobs(limit, offset string, conn *sql.DB) ([]Job, error) {
	query := `SELECT * FROM job LIMIT ? OFFSET ?`
	var err error
	stmt, err := conn.Prepare(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	var rows *sql.Rows
	rows, err = stmt.Query(limit, offset)

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

// Get job and employer detail
func GetJobFromDB(jobID int, conn *sql.DB) (interface{}, error) {
	query := `SELECT jb.id, employer_id,
					jb.title, description,
					responsibility, jb.skills,
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

// Get job detail
func GetJobDetailFromDB(jobID int, conn *sql.DB) (*Job, error) {
	query := `SELECT title, description,
	 			location, employment_type,
				price_from, price_to
			FROM job
			WHERE id = ?`
	stmt, err := conn.Prepare(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	var job Job
	err = stmt.QueryRow(jobID).Scan(
		&job.Title, &job.Description,
		&job.Location, &job.EmployementType,
		&job.PriceFrom, &job.PriceTo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &job, nil
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

// Get jobs where user has an application or user received an application from one of its posted jobs.
//
// Result depends on `querySentOrReceived` to get either received or sent jobs
func GetAppliedOrReceivedJobsFromDB(querySentOrReceived, limit, offset string, accountID int, conn *sql.DB) ([]Application, error) {
	stmt, err := conn.Prepare(querySentOrReceived)
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
		return false, err
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

func RemoveJob(jobID, employerID int, db *sql.DB) error {
	query := `
		DELETE FROM job WHERE id = ? AND employer_id = ?
		AND id NOT IN (SELECT job_id from job_application where job_id = ?)
	`
	if db != nil {
		stmt, err := db.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(jobID, employerID, jobID)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
	return errors.New("no database connection found")
}

// Get users in the database
func GetAccounts(userID interface{}, limit, offset int, conn *sql.DB) ([]UserDetail, error) {
	query := `SELECT * FROM account WHERE is_verified = 1 AND is_admin = 0 LIMIT ? OFFSET ?`
	var userID_ int
	if val, ok := userID.(int); ok {
		query = `SELECT * FROM account WHERE id != ? AND is_verified = 1 AND is_admin = 0 LIMIT ? OFFSET ? `
		userID_ = val
	}

	stmt, err := conn.Prepare(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	var rows *sql.Rows
	if userID_ > 0 {
		rows, err = stmt.Query(userID_, limit, offset)
	} else {
		rows, err = stmt.Query(limit, offset)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var users []UserDetail
	var t time.Time
	var p string
	for rows.Next() {
		var user UserDetail
		err = rows.Scan(
			&user.AccountId,
			&t,
			&user.Name,
			&user.Email,
			&p,
			&user.Birthdate,
			&user.Address,
			&user.IsVerified,
			&user.IsVerificationPending,
			&user.IsAdmin,
			&user.Detail,
			&user.Contact,
			&user.ProfileImage,
			&user.GovIDImage,
			&user.IncomeTaxReturnFile,
			&user.Title,
			&user.Skills,
		)
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

// Updates the user profile details
//
// Fields that can be updated regardless if the user is verified or not
func UpdateUserProfDetails(userID int, skills, detail, title string, conn *sql.DB) error {
	if conn != nil {
		updateStmt := `UPDATE account SET
							skills=?,
							detail=?,
							title=?
				   		WHERE id=?`
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
		_, err = stmt.Exec(skills, detail, title, userID)
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

// Creates a new job record and a proposal record
func CreateAJobProposal(employeeID, employerID int, nj NewJobProposal, conn *sql.DB) error {
	if conn != nil {
		insertJobStmt := `
			INSERT INTO job(dateline, title, description,
					responsibility, skills, location,
					price_from, price_to, employment_type, employer_id)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		insertPrpStmt := `
				INSERT INTO job_proposal(employee_id, employer_id, job_id)
				VALUES(?, ?, ?)
		`
		tx, err := conn.Begin()
		if err != nil {
			log.Println(err)
		}
		stmtJob, err := tx.Prepare(insertJobStmt)
		if err != nil {
			log.Println(err)
			return err
		}

		stmtPrp, err := tx.Prepare(insertPrpStmt)
		if err != nil {
			log.Println(err)
			return err
		}
		defer stmtJob.Close()
		defer stmtPrp.Close()

		row, err := stmtJob.Exec(nj.DateLine, nj.Title,
			nj.Description, nj.ResponsibilitiesToDB,
			nj.SkillsToDB, nj.Location, nj.SalaryRangeFrom,
			nj.SalaryRangeTo, nj.EmploymentType, employerID)
		if err != nil {
			log.Println(err)
			return err
		}

		jobID, err := row.LastInsertId()
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = stmtPrp.Exec(employeeID, employerID, jobID)
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

// Get proposals sent
func GetProposalsFromDB(limit, offset string, employerID int, conn *sql.DB) ([]Proposal, error) {
	if conn != nil {
		query := `SELECT id, employee_id, employer_id,
						 job_id, status
					FROM job_proposal WHERE employer_id = ? 
					LIMIT ? OFFSET ?
				`
		stmt, err := conn.Prepare(query)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer stmt.Close()

		rows, err := stmt.Query(employerID, limit, offset)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer rows.Close()

		var proposals []Proposal
		for rows.Next() {
			var proposal Proposal
			err = rows.Scan(&proposal.ID,
				&proposal.EmployeeID,
				&proposal.EmployerID,
				&proposal.JobID, &proposal.Status)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			proposals = append(proposals, proposal)
		}

		err = rows.Err()
		if err != nil {
			log.Println(err)
			return nil, err
		}

		return proposals, nil
	}
	return nil, errors.New("no db found")
}

// Get proposals received
func GetReceivedProposalsFromDB(limit, offset string, employeeID int, conn *sql.DB) ([]Proposal, error) {
	if conn != nil {
		query := `SELECT id, employee_id, employer_id, job_id, status
					FROM job_proposal WHERE employee_id = ? AND status = 'PENDING'
					LIMIT ? OFFSET ?
				`
		stmt, err := conn.Prepare(query)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer stmt.Close()

		rows, err := stmt.Query(employeeID, limit, offset)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer rows.Close()

		var proposals []Proposal
		for rows.Next() {
			var proposal Proposal
			err = rows.Scan(&proposal.ID,
				&proposal.EmployeeID,
				&proposal.EmployerID,
				&proposal.JobID, &proposal.Status)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			proposals = append(proposals, proposal)
		}

		err = rows.Err()
		if err != nil {
			log.Println(err)
			return nil, err
		}

		return proposals, nil
	}
	return nil, errors.New("no db found")
}

// Update job proposal status
func UpdateJobProposalStatus(action string, proposalID int, conn *sql.DB) error {
	if conn != nil {
		updateStmt := `UPDATE job_proposal SET status = ? WHERE id = ?`
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

		var status string
		if action == "accept" {
			status = "ACCEPTED"
		} else {
			status = "REJECTED"
		}

		_, err = stmt.Exec(status, proposalID)
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

// Update application status
func UpdateJobApplicationStatus(action string, applicationID int, conn *sql.DB) error {
	if conn != nil {
		updateStmt := `UPDATE job_application SET status = ? WHERE id = ?`
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

		var status string
		if action == "accept" {
			status = "ACCEPTED"
		} else {
			status = "REJECTED"
		}

		_, err = stmt.Exec(status, applicationID)
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

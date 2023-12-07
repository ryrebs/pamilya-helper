package db

import (
	"database/sql"
	"fmt"
	"log"
	"mime/multipart"
	"pamilyahelper/webapp/server/utils"
	"strings"
	"time"
)

const (
	UploadDst = "public/uploads" // Upload destination
)

type Job struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Skills          string `json:"skills"`
	Location        string `json:"location"`
	PriceFrom       string `json:"price_from"`
	PriceTo         string `json:"price_to"`
	Responsibility  string `json:"responsibility"`
	EmployementType string `json:"employement_type"`
	EmployerId      int    `json:"employer_id"`
	DateLine        string `json:"dateline"`
	EmployerName    string
	EmployerEmail   string
	EmployerContact string
	EmployerDetail  string
}

type Application struct {
	Job
	Status        string
	ApplicationId int
	Timestamp     time.Time
	EmployeeId    int
	JobID         int
}

type UserDetail struct {
	Name                  string `json:"name"`
	Email                 string `json:"email"`
	Birthdate             string `json:"birthdate"`
	Address               string `json:"address"`
	IsVerified            bool   `json:"is_verified"`
	IsAdmin               bool   `json:"is_admin"`
	IsVerificationPending bool   `json:"is_verification_pending"`
	AccountId             int
	Detail                string `json:"detail"`
	Contact               string `json:"contact"`
	GovIDImage            string
	ProfileImage          string
	Title                 string
	Skills                string
}

type User struct {
	UserDetail
	Password string `json:"password"`
}

type UserVerification struct {
	Name      string
	Email     string
	Birthdate string
	Address   string
	GovId     string
}

type NewUser struct {
	Name     string `form:"name" validate:"required,min=4"`
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required,min=8"`
}

type EditableUserFields struct {
	Name      string `form:"name" validate:"required,min=4"`
	Address   string `form:"address" validate:"required"`
	Birthdate string `form:"birthdate" validate:"required"`
}

func CreateUser(user NewUser, db *sql.DB) error {
	u := User{
		UserDetail: UserDetail{
			Name:  user.Name,
			Email: user.Email,
		},
		Password: user.Password,
	}
	return InsertUser(
		u,
		db,
	)
}

// Returns true is user is valid else false
func ValidateUser(password_text, user_password string) bool {
	return IsPWDvalid(password_text, user_password)
}

// Returns user detail with type UserDetail if found
func FindUserDetail(email string, db *sql.DB) *UserDetail {
	if user := FindUser(email, db); user != (User{}) {
		return &user.UserDetail
	}
	return nil
}

// Returns the user with type User if found
func FindUser(email string, db *sql.DB) User {
	return FindUserFromDb(email, db)
}

// Update user details
func UpdateUserDetail(user UserDetail, details EditableUserFields, file *multipart.FileHeader, db *sql.DB) error {
	newFileName := fmt.Sprint(user.AccountId) + "_" + "gov_id" + file.Filename
	pathFname := fmt.Sprintf(UploadDst + "/" + newFileName)

	err := utils.CreateFile(file, pathFname)
	if err != nil {
		log.Println(err)
		return err
	}

	// Remove created file if any error occurs.
	err = UpdateUser(newFileName, user.Email, details, db)
	if err != nil {
		utils.RemoveFile(pathFname)
		return err
	}
	return nil
}

func UpdateUserDetailProfileImage(userID int, file *multipart.FileHeader, db *sql.DB) error {
	newFileName := fmt.Sprint(userID) + "_" + "profile_" + file.Filename
	pathFname := fmt.Sprintf(UploadDst + "/" + newFileName)

	err := utils.CreateFile(file, pathFname)
	if err != nil {
		log.Println(err)
		return err
	}

	// Remove created file if any error occurs.
	err = UpdateProfileImage(userID, newFileName, db)
	if err != nil {
		utils.RemoveFile(pathFname)
		return err
	}
	return nil
}

func GetAccountsForVerification(limit, offset string, db *sql.DB) ([]UserVerification, error) {
	return GetAccountsForVerificationFromDb(limit, offset, db)
}

func createJobData(jobs []Job) []map[string]interface{} {
	respJobs := []map[string]interface{}{}
	for _, job := range jobs {
		resp := strings.Split(job.Responsibility, "|")
		skills := strings.Split(job.Skills, "|")
		respJobs = append(respJobs, map[string]interface{}{
			"responsibilities": resp,
			"skills":           skills,
			"employer_id":      job.EmployerId,
			"id":               job.ID,
			"title":            job.Title,
			"employment_type":  job.EmployementType,
			"location":         job.Location,
			"description":      job.Description,
			"price_from":       job.PriceFrom,
			"price_to":         job.PriceTo,
			"dateline":         job.DateLine,
		})
	}
	return respJobs
}

// Return jobs not posted by the user
func GetJobs(limit, offset string, accountID int, conn *sql.DB) (interface{}, error) {
	jobs, err := GetJobsFromDB("", limit, offset, accountID, conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	respJobs := createJobData(jobs)
	return respJobs, nil
}

// Get user created jobs
func GetOwnedJobs(limit, offset string, accountID int, conn *sql.DB) (interface{}, error) {
	query := `SELECT * FROM job WHERE employer_id == ? LIMIT ? OFFSET ?`
	jobs, err := GetJobsFromDB(query, limit, offset, accountID, conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	respJobs := createJobData(jobs)
	return respJobs, nil
}

// Get all jobs regardless the owner. This is for anon users.
func GetAllJobs(limit, offset string, conn *sql.DB) (interface{}, error) {
	jobs, err := GetJobsAllJobs(limit, offset, conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	respJobs := createJobData(jobs)
	return respJobs, nil

}

// Get user applied jobs
func GetAppliedJobs(limit, offset string, accountID int, conn *sql.DB) (interface{}, error) {
	jobs, err := GetAppliedJobsFromDB(limit, offset, accountID, conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	respJobs := []map[string]interface{}{}
	for _, job := range jobs {
		resp := strings.Split(job.Responsibility, "|")
		skills := strings.Split(job.Skills, "|")
		respJobs = append(respJobs, map[string]interface{}{
			"responsibilities": resp,
			"skills":           skills,
			"employer_id":      job.EmployerId,
			"id":               job.ID,
			"title":            job.Title,
			"employment_type":  job.EmployementType,
			"location":         job.Location,
			"description":      job.Description,
			"price_from":       job.PriceFrom,
			"price_to":         job.PriceTo,
			"dateline":         job.DateLine,
			"status":           job.Status,
		})
	}
	return respJobs, nil
}

// Get job details
func GetJob(jobID int, conn *sql.DB) (interface{}, error) {
	job, err := GetJobFromDB(jobID, conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	resp := strings.Split(job.(Job).Responsibility, "|")
	skills := strings.Split(job.(Job).Skills, "|")
	data := map[string]interface{}{
		"responsibilities": resp,
		"skills":           skills,
		"employer_id":      job.(Job).EmployerId,
		"id":               job.(Job).ID,
		"title":            job.(Job).Title,
		"employment_type":  job.(Job).EmployementType,
		"location":         job.(Job).Location,
		"description":      job.(Job).Description,
		"price_from":       job.(Job).PriceFrom,
		"price_to":         job.(Job).PriceTo,
		"dateline":         job.(Job).DateLine,
		"employer":         job.(Job).EmployerName,
		"emp_email":        job.(Job).EmployerEmail,
		"emp_contact":      job.(Job).EmployerContact,
		"emp_detail":       job.(Job).EmployerDetail,
	}
	return data, nil
}

func CreateJobApplication(jobID, employeeID int, conn *sql.DB) error {
	return InsertJobApplication(jobID, employeeID, conn)
}

// User data for frontend consumption.
func createUserData(users []UserDetail) []map[string]interface{} {
	var data []map[string]interface{}
	for _, u := range users {
		skills := strings.Split(u.Skills, "|")
		data = append(data, map[string]interface{}{
			"user_id":       u.AccountId,
			"name":          u.Name,
			"title":         u.Title,
			"birthdate":     u.Birthdate,
			"email":         u.Email,
			"address":       u.Address,
			"contact":       u.Contact,
			"detail":        u.Detail,
			"profile_image": u.ProfileImage,
			"skills":        skills,
		})
	}
	return data
}

// Get all users
func GetAllAccountAsAnon(limit, offset int, conn *sql.DB) ([]map[string]interface{}, error) {
	users, err := GetAccounts(nil, limit, offset, conn)
	if err != nil {
		return nil, err
	}
	return createUserData(users), nil
}

// Get all users except current user in session
func GetAllAccountAsUser(userID interface{}, limit, offset int, conn *sql.DB) ([]map[string]interface{}, error) {
	users, err := GetAccounts(userID, limit, offset, conn)
	if err != nil {
		return nil, err
	}
	return createUserData(users), nil
}

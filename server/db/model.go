package db

import (
	"database/sql"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"
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

func FindUserDetail(email string, db *sql.DB) *UserDetail {
	if user := FindUser(email, db); user != (User{}) {
		return &user.UserDetail
	}
	return nil
}

// Returns the user if found
func FindUser(email string, db *sql.DB) User {
	return FindUserFromDb(email, db)
}

// Update user details
func UpdateUserDetail(user UserDetail, details EditableUserFields, file *multipart.FileHeader, db *sql.DB) error {
	rmFFn := func(fname string) error {
		if _, err := os.Stat(fname); err != nil {
			log.Println(err)
			return err
		}
		err := os.Remove(fname)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}

	newFileName := fmt.Sprint(user.AccountId) + "_" + file.Filename
	pathFname := fmt.Sprintf(UploadDst + "/" + newFileName)
	err := CreateFile(file, pathFname, user.AccountId)

	if err != nil {
		log.Println(err)
		return err
	}

	// Remove created file if any error occurs.
	err = UpdateUser(user.Email, details, db)
	if err != nil {
		rmFFn(pathFname)
		return err
	}

	err = InsertGovID(user.AccountId, newFileName, db)
	if err != nil {
		rmFFn(pathFname)
		return err
	}

	return nil
}

func GetAccountsForVerification(limit, offset string, db *sql.DB) ([]UserVerification, error) {
	return GetAccountsForVerificationFromDb(limit, offset, db)
}

func GetJobs(limit, offset string, accountID int, conn *sql.DB) (interface{}, error) {
	jobs, err := GetJobsFromDB(limit, offset, accountID, conn)
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
		})
	}
	return respJobs, nil
}

func GetJob(jobID, empID int, conn *sql.DB) (interface{}, error) {
	job, err := GetJobFromDB(jobID, empID, conn)
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
	}
	return data, nil
}

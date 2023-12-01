package db

import (
	"database/sql"
	"fmt"
	"log"
	"mime/multipart"
	"os"
)

const UploadDst = "public/uploads"

type UserDetail struct {
	Name                  string
	Email                 string
	Birthdate             sql.NullString
	Address               sql.NullString
	IsVerified            bool
	IsAdmin               bool
	IsVerificationPending bool
	AccountId             int
}

type User struct {
	UserDetail
	Password string
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
	return InsertUser(
		user.Name, user.Email,
		user.Password,
		false,
		false,
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
	pathFname := fmt.Sprintf(UploadDst + "/" + fmt.Sprint(user.AccountId) + "_" + file.Filename)
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

	err = InsertGovID(user.AccountId, file.Filename, db)
	if err != nil {
		rmFFn(pathFname)
		return err
	}

	return nil
}

package db

import (
	"database/sql"
	"log"
)

type UserDetail struct {
	Name        string
	Email       string
	Birthdate   sql.NullString
	Address     sql.NullString
	Is_verified bool
	Is_admin    bool
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
	Address   string `form:"email" validate:"required"`
	Birthdate string `form:"email" validate:"required"`
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
	stmt, err := db.Prepare(`SELECT name, email,
								password, birthdate,
								address, is_verified, is_admin
							FROM account WHERE email=?`)
	if err != nil {
		log.Fatal(err)
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
		&user.Is_verified,
		&user.Is_admin)

	if err != nil {
		log.Println(err)
		return User{}
	}
	return user
}

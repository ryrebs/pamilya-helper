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

func validateUser(password_text string, user User) UserDetail {
	if IsPWDvalid(password_text, user.Password) {
		return UserDetail{
			Name:        user.Name,
			Email:       user.Email,
			Birthdate:   user.Birthdate,
			Address:     user.Address,
			Is_verified: user.Is_verified,
			Is_admin:    user.Is_admin,
		}

	}
	return UserDetail{}
}

func FindUser(email, password string, db *sql.DB) UserDetail {
	stmt, err := db.Prepare(`SELECT name, email,
								password, birthdate,
								address, is_verified, is_admin
							FROM account WHERE email=?`)
	if err != nil {
		log.Fatal(err)
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
		return UserDetail{}
	}
	return validateUser(password, user)
}

package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type User struct {
	name        string
	email       string
	password    string
	birthdate   time.Time
	address     sql.NullString
	id_type     sql.NullString
	is_verified bool
	is_admin    bool
}

type UserPublicData struct {
	Name        string
	Email       string
	Birthdate   time.Time
	Address     string
	Id_type     string
	Is_verified bool
	Is_admin    bool
}

func CreateUser(user *User) {
	theTime := time.Date(2021, 8, 15, 14, 0, 0, 0, time.Local)
	fmt.Println("The time is", theTime)
	y, m, d := theTime.Date()
	dt := fmt.Sprintf("%d-%d-%d", y, m, d)
	fmt.Print(dt)
}

func validateUser(password_text string, user User) UserPublicData {
	if IsPWDvalid(password_text, user.password) {
		return UserPublicData{
			Name:        user.name,
			Email:       user.email,
			Birthdate:   user.birthdate,
			Address:     user.address.String,
			Id_type:     user.id_type.String,
			Is_verified: user.is_verified,
			Is_admin:    user.is_admin,
		}

	}
	return UserPublicData{}
}

func FindUser(email, password string, db *sql.DB) UserPublicData {
	stmt, err := db.Prepare(`SELECT name, email,
								password, birthdate,
								address, id_type, is_verified, is_admin
							FROM account WHERE email=?`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var user User
	err = stmt.QueryRow(email).Scan(
		&user.name,
		&user.email,
		&user.password,
		&user.birthdate,
		&user.address,
		&user.id_type,
		&user.is_verified,
		&user.is_admin)

	if err != nil {
		log.Println(err)
		return UserPublicData{}
	}
	return validateUser(password, user)
}

func SignInUser(name, password string) {}

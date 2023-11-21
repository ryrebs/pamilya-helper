package db

type User struct {
	name     string
	email    string
	password string
	is_admin bool
}

func CreateUser(user *User) {}

func FindUser(name, email string) {

}

func SignInUser(name, password string) {}

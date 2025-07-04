package routes

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const PMHSessionName = "auth-pamilyahelper-session"

// Given a session exists for the user,
// this function should extract the user details from db.
func GetUserFromSession(c echo.Context, db_ *sql.DB) (*db.UserDetail, error) {
	sess, _ := session.Get(PMHSessionName, c)

	// Session not found
	if sess != nil && sess.Values["user"] != nil {
		if user := db.FindUserDetail(sess.Values["user"].(string), db_); user != nil {
			return user, nil
		}
	}
	return nil, errors.New("no user found for this session")
}

func createSession(userEmail string, c *db.CustomDBContext) error {
	sess, _ := session.Get(PMHSessionName, c)
	sess.Options = &sessions.Options{
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	sess.Values["user"] = userEmail
	return sess.Save(c.Request(), c.Response())
}

func SignIn(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	r := cc.Request()

	if r.Method == "GET" {
		return cc.Render(http.StatusOK, "signin-signup.html", nil)
	}

	if r.Method == "POST" {
		email := cc.FormValue("email")
		password := cc.FormValue("password")
		user := db.FindUser(email, cc.Db())

		// No user found
		if (db.User{}) == user {
			return cc.Render(http.StatusOK, "signin-signup.html", map[string]interface{}{
				"msg": "Invalid Email or Password.",
			})

		}

		// User found;validate
		if !db.ValidateUser(password, user.Password) {
			return cc.Render(http.StatusOK, "signin-signup.html", map[string]interface{}{
				"msg": "Invalid Email or Password.",
			})
		}

		// User found. Set session cookies
		err := createSession(user.Email, cc)
		if err != nil {
			log.Println(err)
			return cc.Render(http.StatusOK, "signin-signup.html", map[string]interface{}{
				"msg": "Invalid Email or Password.",
			})
		}
		return cc.Redirect(http.StatusSeeOther, "/")
	}
	return cc.Render(http.StatusMethodNotAllowed, "signin-signup.html", nil)
}

func SignOut(c echo.Context) error {
	sess, _ := session.Get(PMHSessionName, c)
	if sess != nil && sess.Values["user"] != nil {
		sess.Options = &sessions.Options{
			MaxAge:   -1,
			Path:     "/",
			HttpOnly: true,
		}
		err := sess.Save(c.Request(), c.Response())
		if err != nil {
			log.Println(err)
			return c.Render(http.StatusInternalServerError, "/users/profile", map[string]interface{}{
				"msg": "Please try again later.",
			})
		}
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

func SignUp(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	user := new(db.NewUser)

	if err := cc.Bind(user); err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)

	}
	if err := cc.Validate(user); err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Create User if user not exists.
	existing_user := db.FindUser(user.Email, cc.Db())
	if (existing_user == db.User{}) {
		if err := db.CreateUser(*user, cc.Db()); err != nil {
			return c.Render(http.StatusBadRequest, "signin-signup.html", map[string]interface{}{
				"msg_signup": "Invalid user mail. Please try again later.",
			})
		}
		// Set session
		err := createSession(user.Email, cc)
		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		// Redirect Index
		return c.Redirect(http.StatusSeeOther, "/")
	}

	return c.Render(http.StatusBadRequest, "signin-signup.html", map[string]interface{}{
		"msg_signup": "Invalid user e-mail. Please try again later.",
	})

}

// Remove user from db. Use only
// for testing purposes or privileged access
func RemoveUser(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	user := new(struct {
		Email string `json:"email" validate:"required,email"`
	})

	if err := cc.Bind(user); err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)

	}
	if err := cc.Validate(user); err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := db.RemoveUser(user.Email, cc.Db())
	if err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())

	}
	return cc.NoContent(http.StatusNoContent)
}

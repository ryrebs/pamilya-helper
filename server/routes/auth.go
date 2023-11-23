package routes

import (
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

var validate *validator.Validate

func RedirectIfSigned(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("auth-pamilyahelper-session", c)
		if sess != nil && sess.Values["user"] != nil {
			c.Redirect(http.StatusSeeOther, "/users/profile")
		}
		return next(c)
	}
}

func createSession(userEmail string, c *db.CustomDBContext) error {
	sess, _ := session.Get("auth-pamilyahelper-session", c)
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

		user := db.FindUser(email, password, cc.Db())

		if (db.UserDetail{}) == user {
			return cc.Render(http.StatusOK, "signin-signup.html", map[string]interface{}{
				"msg": "Invalid Email or Password.",
			})

		}

		// User found. Set session cookies
		err := createSession(user.Name, cc)
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
	sess, _ := session.Get("auth-pamilyahelper-session", c)
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
	validate = validator.New()

	if err := cc.Bind(user); err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)

	}
	if err := validate.Struct(user); err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	// Create User
	if err := db.CreateUser(*user, cc.Db()); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
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

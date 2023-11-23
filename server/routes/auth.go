package routes

import (
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func RedirectIfSigned(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("auth-pamilyahelper-session", c)
		if sess != nil && sess.Values["user"] != nil {
			c.Redirect(http.StatusSeeOther, "/users/profile")
		}
		return next(c)
	}
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

		if (db.UserPublicData{}) == user {
			return cc.Render(http.StatusOK, "signin-signup.html", map[string]interface{}{
				"msg": "Invalid Email or Password.",
			})

		}
		// User found. Set session cookies
		sess, _ := session.Get("auth-pamilyahelper-session", cc)
		sess.Options = &sessions.Options{
			MaxAge:   86400 * 7,
			HttpOnly: true,
		}
		sess.Values["user"] = user.Name
		err := sess.Save(cc.Request(), cc.Response())
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

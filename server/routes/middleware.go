package routes

import (
	"net/http"
	"pamilyahelper/webapp/server/db"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Redirects the user to profile if user is signed in.
func RedirectToProfileMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("auth-pamilyahelper-session", c)
		cc := c.(*db.CustomDBContext)

		if sess != nil && sess.Values["user"] != nil {
			if user := db.FindUser(sess.Values["user"].(string), cc.Db()); user != (db.User{}) {
				return c.Redirect(http.StatusSeeOther, "/users/profile")
			}
		}
		return next(c)
	}
}

// Redirects the user to signin if user is not signin in.
func RequireSignInMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("auth-pamilyahelper-session", c)
		cc := c.(*db.CustomDBContext)

		err := CheckSessionExist(c, sess, cc.Db())
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/signin")
		}
		return next(c)
	}
}

// Requires route to be admin accessible only.
func RequireAdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("auth-pamilyahelper-session", c)
		cc := c.(*db.CustomDBContext)

		err := CheckSessionIsAdmin(c, sess, cc.Db())
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/signin")
		}
		return next(c)
	}
}

// Requires route to be admin accessible only.
func RequireVerifiedUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("auth-pamilyahelper-session", c)
		cc := c.(*db.CustomDBContext)

		if v, _ := CheckUserIsVerified(c, sess, cc.Db()); v {
			return next(c)
		}
		return c.Redirect(http.StatusSeeOther, "/signin")
	}
}

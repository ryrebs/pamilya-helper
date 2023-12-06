package routes

import (
	"database/sql"
	"errors"
	"pamilyahelper/webapp/server/db"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

func CheckUserIsVerified(c echo.Context, sess *sessions.Session, conn *sql.DB) (bool, error) {
	// Session not found
	if sess != nil && sess.Values["user"] == nil {
		return false, errors.New("no session found")
	}

	user := db.FindUser(sess.Values["user"].(string), conn)

	// Session found but email not found
	if sess.Values["user"] != nil && user == (db.User{}) {
		return false, errors.New("no user found")
	}

	// Check user is verified
	if user.IsVerified {
		return true, nil
	}

	return false, errors.New("no user found")
}

func CheckSessionExist(c echo.Context, sess *sessions.Session, conn *sql.DB) error {
	// Session not found
	if sess != nil && sess.Values["user"] == nil {
		return errors.New("no session found")
	}
	// Session found but email not found
	if sess.Values["user"] != nil && db.FindUser(sess.Values["user"].(string), conn) == (db.User{}) {
		return errors.New("no user found")
	}
	return nil
}

func CheckSessionIsAdmin(c echo.Context, sess *sessions.Session, conn *sql.DB) error {
	// Session not found
	if sess != nil && sess.Values["user"] == nil {
		return errors.New("no session found")
	}

	user := db.FindUser(sess.Values["user"].(string), conn)

	// Session found but email not found
	if sess.Values["user"] != nil && user == (db.User{}) {
		return errors.New("no user found")
	}

	// Check admin
	if !user.IsAdmin {
		return errors.New("admin not found")
	}

	return nil
}

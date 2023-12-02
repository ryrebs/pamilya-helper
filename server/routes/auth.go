package routes

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Given a session exists for the user,
// this function should extract the user details from db.
func GetUserFromSession(c echo.Context, db_ *sql.DB) (*db.UserDetail, error) {
	sess, _ := session.Get("auth-pamilyahelper-session", c)

	// Session not found
	if sess != nil && sess.Values["user"] != nil {
		if user := db.FindUserDetail(sess.Values["user"].(string), db_); user != nil {
			return user, nil
		}
	}
	return nil, errors.New("no user found for this session")
}

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

// Verify account view
func VerifyAccountView(c echo.Context) error {
	cc := c.(*db.CustomDBContext)

	data := map[string]interface{}{
		"is_log_in": true,
	}
	user, err := GetUserFromSession(cc, cc.Db())
	if user != nil {
		govIdFile := db.GetUserGovId(user.AccountId, cc.Db())
		data["data"] = map[string]interface{}{
			"name":                    user.Name,
			"email":                   user.Email,
			"birthdate":               user.Birthdate.String,
			"address":                 user.Address.String,
			"is_admin":                user.IsAdmin,
			"is_verified":             user.IsVerified,
			"is_verification_pending": user.IsVerificationPending,
			"gov_id":                  govIdFile,
		}
	}
	if err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// Handle Get requests
	if c.Request().Method == "GET" {
		return renderWithAuthContext("verify-profile.html", c, data)
	}

	errorMsgs := make([]string, 0, 4)
	editableUser := db.EditableUserFields{}

	if err := cc.Bind(&editableUser); err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)

	}
	if err := cc.Validate(editableUser); err != nil {
		log.Println(err.Error())
		m := strings.Split(err.Error(), "\n")
		for _, v := range m {
			if strings.Contains(v, "EditableUserFields.Name") {
				errorMsgs = append(errorMsgs, "Invalid Name")
			}
			if strings.Contains(v, "EditableUserFields.Birthdate") {
				errorMsgs = append(errorMsgs, "Invalid Birthdate")
			}
			if strings.Contains(v, "EditableUserFields.Address") {
				errorMsgs = append(errorMsgs, "Invalid Address")
			}
		}
		data["msgs"] = errorMsgs
		return renderWithAuthContext("verify-profile.html", cc, data)
	}
	file, err := cc.FormFile("file")
	if err != nil {
		log.Println(err.Error())
		errorMsgs = append(errorMsgs, "Invalid File")
		data["msgs"] = errorMsgs
		return renderWithAuthContext("verify-profile.html", cc, data)
	}

	// Update user details
	err = db.UpdateUserDetail(*user, editableUser, file, cc.Db())
	if err != nil {
		log.Println(err.Error())
		data["msgs"] = "Something went wrong. Please try again later."
		return renderWithAuthContext("verify-profile.html", cc, data)
	}
	data["success_msg"] = "Success. Waiting for approval."
	data["data"] = map[string]interface{}{
		"name":                    editableUser.Name,
		"email":                   user.Email,
		"birthdate":               editableUser.Birthdate,
		"address":                 editableUser.Address,
		"is_admin":                user.IsAdmin,
		"is_verified":             user.IsVerified,
		"is_verification_pending": user.IsVerificationPending,
		"gov_id":                  file.Filename,
	}
	return renderWithAuthContext("verify-profile.html", cc, data)
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

func VerifyUser(c echo.Context) error {
	user := struct {
		action string `form:"name" validate="required"`
		email  string `form:"email" validated="required,oneof=accept reject"`
	}{}
	log.Println(user)
	return nil
}

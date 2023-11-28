package routes

import (
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func renderWithAuthContext(templateName string, c echo.Context, data interface{}) error {
	sess, _ := session.Get("auth-pamilyahelper-session", c)
	if sess != nil && sess.Values["user"] != nil {
		return c.Render(http.StatusOK, templateName, data)
	}
	return c.Render(http.StatusOK, templateName, nil)
}

func Index(c echo.Context) error {
	return renderWithAuthContext("index.html", c, map[string]interface{}{
		"is_log_in": true,
	})
}

func About(c echo.Context) error {
	return c.Render(http.StatusOK, "about.html", nil)
}

func Profile(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	data := map[string]interface{}{
		"is_log_in": true,
	}

	if cc.Request().Method == "PATCH" {
		name := cc.FormValue("name")
		birtdate := cc.FormValue("birthdate")
		address := cc.FormValue("address")
		errorMsgs := ""
		validate = validator.New()

		if err := validate.Struct(db.EditableUserFields{
			Name:      name,
			Birthdate: birtdate,
			Address:   address,
		}); err != nil {
			m := strings.Split(err.Error(), "\n")
			for _, v := range m {
				if strings.Contains(v, "EditableUserFields.Name") {
					errorMsgs = errorMsgs + "Invalid Name\n"
				}
				if strings.Contains(v, "EditableUserFields.Birthdate") {
					errorMsgs = errorMsgs + "Invalid Birthdate\n"
				}
				if strings.Contains(v, "EditableUserFields.Address") {
					errorMsgs = errorMsgs + "Invalid Address\n"
				}
			}
			data["msgs"] = errorMsgs
			return renderWithAuthContext("profile.html", cc, data)
		}

		file, err := c.FormFile("file")
		if err != nil {
			return nil
		}
		log.Println(file.Filename)

		// Handle save file and update account details
		// If File exists remove file and save new file for Gov ID
		// Save file first before adding/delete/updating entry on db
		// Duplicate file for non GOV ID should be appended by timestamp for uniqueness
		// Detail column is added on upload for additional meta data.
	}

	user, error := GetUserFromSession(c)

	// Return account details on GET request
	if user != nil {

		data["data"] = map[string]interface{}{
			"email":       user.Email,
			"birthdate":   user.Birthdate.String,
			"address":     user.Address.String,
			"is_admin":    user.Is_admin,
			"is_verified": user.Is_verified,
		}
		return renderWithAuthContext("profile.html", c, data)
	}

	log.Println(error)
	return echo.NewHTTPError(http.StatusInternalServerError)
}

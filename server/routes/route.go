package routes

import (
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	user, error := GetUserFromSession(c, cc.Db())
	if user != nil {
		govIdFile := db.GetUserGovId(user.AccountId, cc.Db())
		data["data"] = map[string]interface{}{
			"name":        cases.Title(language.English, cases.Compact).String(user.Name),
			"email":       user.Email,
			"birthdate":   user.Birthdate.String,
			"address":     user.Address.String,
			"is_admin":    user.IsAdmin,
			"is_verified": user.IsVerified,
			"gov_id":      govIdFile,
		}
		return renderWithAuthContext("profile.html", c, data)
	}

	log.Println(error)
	return echo.NewHTTPError(http.StatusInternalServerError)
}

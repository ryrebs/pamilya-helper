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

// Add field 'is_log_in' for frontend templates if a user is found
func renderWithAuthContext(templateName string, c echo.Context, data map[string]interface{}) error {
	if data == nil {
		data = map[string]interface{}{}
	}
	sess, _ := session.Get("auth-pamilyahelper-session", c)
	if sess != nil && sess.Values["user"] != nil {
		data["is_log_in"] = true
		return c.Render(http.StatusOK, templateName, data)
	}
	return c.Render(http.StatusOK, templateName, data)
}

func Index(c echo.Context) error {
	return renderWithAuthContext("index.html", c, map[string]interface{}{})
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
	var profile_data map[string]interface{}
	var applications []map[string]interface{}

	if user != nil {
		// Get what type of info should be returned
		infoType := struct {
			info string `query:"info" validate:"oneof=profile applications"`
		}{
			info: "profile",
		}

		err := cc.Bind(&infoType)
		if err != nil {
			log.Println(err)
		}

		// Get accounts with pending verifications
		govIdFile := db.GetUserGovId(user.AccountId, cc.Db())
		var accounts []db.UserVerification
		if user.IsAdmin {
			accounts_, _ := db.GetAccountsForVerification("3", "0", cc.Db())
			accounts = accounts_
		}

		switch infoType.info {
		case "profile":
			profile_data = map[string]interface{}{
				"name":        cases.Title(language.English, cases.Compact).String(user.Name),
				"email":       user.Email,
				"birthdate":   user.Birthdate,
				"address":     user.Address,
				"is_admin":    user.IsAdmin,
				"is_verified": user.IsVerified,

				"gov_id": govIdFile,
			}
		}

		data["data"] = map[string]interface{}{
			"profile":      profile_data,
			"applications": applications,
			"accounts":     accounts,
			"infoType":     infoType,
		}
		return renderWithAuthContext("profile.html", c, data)
	}

	log.Println(error)
	return echo.NewHTTPError(http.StatusInternalServerError)
}

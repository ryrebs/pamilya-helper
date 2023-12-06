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
	cc := c.(*db.CustomDBContext)
	user, err := GetUserFromSession(c, cc.Db())

	// Anon user
	if err != nil {
		jobs, _ := db.GetAllJobs("10", "0", cc.Db())
		return renderWithAuthContext("index.html", cc, map[string]interface{}{
			"jobs": jobs,
		})
	}

	// Logged user
	jobs, err := db.GetJobs("10", "0", user.AccountId, cc.Db())
	if err != nil {
		log.Println(err)
		return renderWithAuthContext("index.html", cc, nil)
	}

	return renderWithAuthContext("index.html", cc, map[string]interface{}{
		"jobs": jobs,
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
	var profile_data map[string]interface{}
	var applications []map[string]interface{}
	var postedJobs []map[string]interface{}

	if user != nil {
		// Get what type of info should be returned
		infoType := struct {
			Info string `query:"info" validate:"oneof=profile applications posted"`
		}{
			Info: "profile",
		}
		err := cc.Bind(&infoType)
		if err != nil {
			log.Println(err)
		}

		// Get accounts with pending verifications
		var accounts []db.UserVerification
		if user.IsAdmin {
			accounts_, _ := db.GetAccountsForVerification("3", "0", cc.Db())
			accounts = accounts_
		}

		switch infoType.Info {
		case "profile":
			profile_data = map[string]interface{}{
				"name":          cases.Title(language.English, cases.Compact).String(user.Name),
				"email":         user.Email,
				"birthdate":     user.Birthdate,
				"address":       user.Address,
				"is_admin":      user.IsAdmin,
				"is_verified":   user.IsVerified,
				"gov_id_image":  user.GovIDImage,
				"profile_image": user.ProfileImage,
			}
		case "applications":
			{
				// Get jobs
				jobs, jErr := db.GetAppliedJobs("10", "0", user.AccountId, cc.Db())
				if jErr != nil {
					log.Println(jErr)
				} else {
					applications = jobs.([]map[string]interface{})
				}
			}
		case "posted":
			{
				// Get jobs
				jobs, jErr := db.GetOwnedJobs("10", "0", user.AccountId, cc.Db())
				if jErr != nil {
					log.Println(jErr)
				} else {
					postedJobs = jobs.([]map[string]interface{})
				}
			}
		}
		data["data"] = map[string]interface{}{
			"profile":      profile_data,
			"applications": applications,
			"postedJobs":   postedJobs,
			"accounts":     accounts,
			"infoType":     infoType.Info,
		}
		return renderWithAuthContext("profile.html", c, data)
	}

	log.Println(error)
	return echo.NewHTTPError(http.StatusInternalServerError)
}

func Contact(c echo.Context) error {
	return renderWithAuthContext("contact.html", c, nil)
}

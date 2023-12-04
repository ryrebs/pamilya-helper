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
	if user != nil {
		// Get accounts with pending verifications
		govIdFile := db.GetUserGovId(user.AccountId, cc.Db())
		var accounts []db.UserVerification
		if user.IsAdmin {
			accounts_, _ := db.GetAccountsForVerification("3", "0", cc.Db())
			accounts = accounts_
		}

		data["data"] = map[string]interface{}{
			"name":        cases.Title(language.English, cases.Compact).String(user.Name),
			"email":       user.Email,
			"birthdate":   user.Birthdate,
			"address":     user.Address,
			"is_admin":    user.IsAdmin,
			"is_verified": user.IsVerified,
			"gov_id":      govIdFile,
			"accounts":    accounts,
		}
		return renderWithAuthContext("profile.html", c, data)
	}

	log.Println(error)
	return echo.NewHTTPError(http.StatusInternalServerError)
}

func JobList(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	data := map[string]interface{}{
		"is_verified": false,
		"jobs":        []db.Job{},
	}
	user, uErr := GetUserFromSession(c, cc.Db())

	// Get jobs
	jobs, jErr := db.GetJobs("10", "0", user.AccountId, cc.Db())
	if jErr != nil {
		log.Println(jErr)
	} else {
		data["jobs"] = jobs
	}

	// If no user return data
	if uErr != nil {
		return c.Render(http.StatusOK, "job-list.html", data)
	}

	// Else set additional fields and return data
	data["is_verified"] = user.IsVerified

	return renderWithAuthContext("job-list.html", c, data)
}

func JobDetail(c echo.Context) error {
	cc := c.(*db.CustomDBContext)

	jDetail := struct {
		ID int `param:"id"`
	}{}
	err := c.Bind(&jDetail)

	if err != nil {
		log.Println(err)
		return renderWithAuthContext(
			"job-detail.html", c, nil,
		)
	}

	user, err := GetUserFromSession(c, cc.Db())
	if err != nil {
		return renderWithAuthContext(
			"job-detail.html", c, nil,
		)
	}

	job, err := db.GetJob(jDetail.ID, user.AccountId, cc.Db())
	if err != nil {
		return renderWithAuthContext(
			"job-detail.html", c, nil,
		)
	}

	return renderWithAuthContext(
		"job-detail.html", c, map[string]interface{}{
			"job": job,
		},
	)
}

package routes

import (
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"
	"pamilyahelper/webapp/server/utils"
	"strings"

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
		jobs, _ := db.GetAllJobs("5", "0", cc.Db())
		users, _ := db.GetAllAccountAsAnon(3, 0, cc.Db())
		return renderWithAuthContext("index.html", cc, map[string]interface{}{
			"jobs":    jobs,
			"helpers": users,
		})
	}

	// Logged user
	jobs, err := db.GetJobs("5", "0", user.AccountId, cc.Db())
	if err != nil {
		log.Println(err)
		return renderWithAuthContext("index.html", cc, nil)
	}

	users, _ := db.GetAllAccountAsUser(user.AccountId, 3, 0, cc.Db())
	return renderWithAuthContext("index.html", cc, map[string]interface{}{
		"jobs":        jobs,
		"helpers":     users,
		"is_verified": user.IsVerified,
	})
}

func About(c echo.Context) error {
	return c.Render(http.StatusOK, "about.html", nil)
}

func Profile(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	user, error := GetUserFromSession(c, cc.Db())
	var profile_data map[string]interface{}
	var applications []map[string]interface{}
	var postedJobs []map[string]interface{}
	var sentProposals []map[string]interface{}
	var receivedProposals []map[string]interface{}
	var receivedApplications []map[string]interface{}

	if user != nil {
		data := map[string]interface{}{
			"is_verified": user.IsVerified,
		}

		// Handle profile updates
		if cc.Request().Method == "POST" {
			var uDt struct {
				Title  string   `form:"title"`
				Skills []string `form:"skills"`
				Detail string   `form:"detail"`
			}
			if err := cc.Bind(&uDt); err != nil {
				log.Println(err)
				return cc.Redirect(http.StatusSeeOther, "/users/profile")
			}
			if err := cc.Validate(uDt); err != nil {
				log.Println(err)
				return cc.Redirect(http.StatusSeeOther, "/users/profile")
			}
			skills := utils.CreateListString(uDt.Skills, nil)
			db.UpdateUserProfDetails(user.AccountId, skills, uDt.Detail, uDt.Title, cc.Db())
			return cc.Redirect(http.StatusSeeOther, "/users/profile")
		}

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
			{
				// Make sure skills is initialized to length 5
				skills := strings.Split(user.Skills, "|")
				if user.Skills == "" {
					skills = make([]string, 5)
				}
				profile_data = map[string]interface{}{
					"name":              cases.Title(language.English, cases.Compact).String(user.Name),
					"email":             user.Email,
					"birthdate":         user.Birthdate,
					"address":           user.Address,
					"is_admin":          user.IsAdmin,
					"gov_id_image":      user.GovIDImage,
					"profile_image":     user.ProfileImage,
					"detail":            user.Detail,
					"title":             user.Title,
					"skills":            skills,
					"income_tax_return": user.IncomeTaxReturnFile,
				}
			}
		case "posted":
			{
				// Get jobs you created
				jobs, jErr := db.GetOwnedJobs("10", "0", user.AccountId, cc.Db())
				if jErr != nil {
					log.Println(jErr)
				} else {
					postedJobs = jobs.([]map[string]interface{})
				}
			}
		case "applications":
			{
				// Get jobs where you sent applications
				jobs, jErr := db.GetAppliedJobs("10", "0", user.AccountId, cc.Db())
				if jErr != nil {
					log.Println(jErr)
				} else {
					applications = jobs.([]map[string]interface{})
				}
			}
		case "rcv_applications":
			{
				// Get received applications
				jobs, jErr := db.GetReceivedApplications("10", "0", user.AccountId, cc.Db())
				if jErr != nil {
					log.Println(jErr)
				} else {
					receivedApplications = jobs.([]map[string]interface{})
				}
			}
		case "proposals":
			{
				// Get proposals you sent to employees
				p, err := db.GetProposals("10", "0", user.AccountId, cc.Db())
				if err != nil {
					log.Println(err)
				} else {
					sentProposals = p
				}
			}
		case "rcv_proposals":
			{
				// Get received proposals from employers
				p, err := db.GetReceviedProposals("10", "0", user.AccountId, cc.Db())
				if err != nil {
					log.Println(err)
				} else {
					receivedProposals = p
				}
			}
		}

		data["data"] = map[string]interface{}{
			"profile":          profile_data,
			"applications":     applications,
			"postedJobs":       postedJobs,
			"proposals":        sentProposals,
			"rcv_proposals":    receivedProposals,
			"rcv_applications": receivedApplications,
			"accounts":         accounts,
			"infoType":         infoType.Info,
		}

		return renderWithAuthContext("profile.html", c, data)
	}

	log.Println(error)
	return echo.NewHTTPError(http.StatusInternalServerError)
}

func Contact(c echo.Context) error {
	return renderWithAuthContext("contact.html", c, nil)
}

func Helper(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	var flashMsg string

	// Check for flash messages
	sess, err := session.Get("post-proposal", cc)
	if err != nil {
		log.Println(err)
	} else {
		v := sess.Flashes("proposal_msg")
		if len(v) >= 1 {
			flashMsg = v[0].(string)
		}
	}

	user, err := GetUserFromSession(c, cc.Db())
	if err != nil {
		users, _ := db.GetAllAccountAsAnon(10, 0, cc.Db())
		return renderWithAuthContext("helpers.html", cc, map[string]interface{}{
			"helpers":  users,
			"flashMsg": flashMsg,
		})
	}

	users, _ := db.GetAllAccountAsUser(user.AccountId, 10, 0, cc.Db())
	return renderWithAuthContext("helpers.html", cc, map[string]interface{}{
		"helpers":  users,
		"flashMsg": flashMsg,
	})
}

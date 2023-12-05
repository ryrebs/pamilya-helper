package routes

import (
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"
	"pamilyahelper/webapp/server/utils"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type CustomJob struct {
	msg string
}

func (c CustomJob) GetMsg() string {
	return c.msg
}

func (c CustomJob) IsMsgError() bool {
	return utils.IsErrorMSG(c.msg)
}

func JobList(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	data := map[string]interface{}{
		"is_verified": false,
		"jobs":        []db.Job{},
	}
	flashMsg := ""
	user, uErr := GetUserFromSession(c, cc.Db())

	// If no user return data
	if uErr != nil {
		return c.Render(http.StatusOK, "job-list.html", data)
	}

	// Get jobs
	jobs, jErr := db.GetJobs("10", "0", user.AccountId, cc.Db())
	if jErr != nil {
		log.Println(jErr)
	} else {
		data["jobs"] = jobs
	}

	// Check for flash messages
	sess, err := session.Get("post-message", cc)
	if err != nil {
		log.Println(err)
	} else {
		for _, v := range sess.Flashes("post_apply") {
			flashMsg = v.(string)
		}
	}

	// Else set additional fields and return data
	data["is_verified"] = user.IsVerified
	data["job_msg"] = CustomJob{
		msg: flashMsg,
	}
	return renderWithAuthContext("job-list.html", c, data)
}

func JobDetail(c echo.Context) error {
	cc := c.(*db.CustomDBContext)

	jDetail := struct {
		ID       int  `param:"id"`
		ViewOnly bool `query:"view"`
	}{}
	err := c.Bind(&jDetail)

	if err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusSeeOther, "/job-list")
	}

	job, err := db.GetJob(jDetail.ID, cc.Db())
	if err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusSeeOther, "/job-list")
	}

	if cc.Request().Method == "POST" {
		// Set flash message
		sess, err := session.Get("post-message", cc)
		if err != nil {
			log.Println(err)
		} else {
			sess.Options = &sessions.Options{
				MaxAge:   10,
				Path:     "/job-list",
				HttpOnly: true,
			}
		}

		user, err := GetUserFromSession(cc, cc.Db())
		if err != nil {
			log.Println(err)
		}

		// Create job application
		err = db.CreateJob(jDetail.ID, user.AccountId, cc.Db())
		if err != nil {
			log.Println(err)
			sess.AddFlash("Something went wrong please try again.", "post_apply")
		} else {
			sess.AddFlash("Application submitted!", "post_apply")
		}

		sess.Save(cc.Request(), cc.Response())
		return cc.Redirect(http.StatusSeeOther, "/job-list")
	}

	return renderWithAuthContext(
		"job-detail.html", c, map[string]interface{}{
			"job":      job,
			"viewOnly": jDetail.ViewOnly,
		},
	)
}

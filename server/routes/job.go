package routes

import (
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"
	"pamilyahelper/webapp/server/utils"
	"strings"

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
		jobs, _ := db.GetAllJobs("20", "0", cc.Db())
		data["jobs"] = jobs
		return c.Render(http.StatusOK, "job-list.html", data)
	}

	// Get jobs
	jobs, jErr := db.GetJobs("20", "0", user.AccountId, cc.Db())
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

	user, userErr := GetUserFromSession(cc, cc.Db())
	is_verified := false
	if userErr != nil {
		log.Println(err)
	} else {
		is_verified = user.IsVerified
	}

	if cc.Request().Method == "POST" {
		// Create session for flash message
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

		// Create job application if user is not the owner of the job
		if user != nil {
			if owned, err := db.IsJobOwned(jDetail.ID, user.AccountId, cc.Db()); !owned && err == nil {
				err = db.CreateJobApplication(jDetail.ID, user.AccountId, cc.Db())
				if err != nil {
					log.Println(err)
					sess.AddFlash("Something went wrong please try again.", "post_apply")
				} else {
					sess.AddFlash("Application submitted!", "post_apply")
				}
			} else {
				log.Println(err)
				sess.AddFlash("Something went wrong please try again.", "post_apply")
			}
			sess.Save(cc.Request(), cc.Response())
		}
		return cc.Redirect(http.StatusSeeOther, "/job-list")
	}

	return renderWithAuthContext(
		"job-detail.html", c, map[string]interface{}{
			"job":         job,
			"view_only":   jDetail.ViewOnly,
			"is_verified": is_verified,
		},
	)
}

func CreateJob(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	user, err := GetUserFromSession(c, cc.Db())
	newJob := struct {
		DateLine         string   `form:"dateline" validate:"required"`
		Title            string   `form:"title" validate:"required"`
		Skills           []string `form:"skills" validate:"required"`
		Responsibilities []string `form:"responsibilities" validate:"required"`
		Description      string   `form:"description" validate:"required"`
		SalaryRangeFrom  string   `form:"salary_range1" validate:"required"`
		SalaryRangeTo    string   `form:"salary_range2" validate:"required"`
		Address          string   `form:"address" validate:"required"`
		EmployementType  string   `form:"employment_type" validate:"required,oneof='Part Time' 'Full Time'"`
	}{
		Skills: make([]string, 0),
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden)
	}
	data := map[string]interface{}{
		"name":      user.Name,
		"email":     user.Email,
		"contact":   user.Contact,
		"detail":    user.Detail,
		"errorMsgs": []string{},
	}

	if cc.Request().Method == "POST" {
		err := cc.Bind(&newJob)
		if err != nil {
			log.Println(err)
		}
		err = cc.Validate(newJob)
		if err != nil {
			data["errorMsgs"] = strings.Split(err.Error(), "\n")
			return renderWithAuthContext("create-job.html", c, data)
		}

		// Create job
		resp := utils.CreateListString(newJob.Responsibilities)
		skills := utils.CreateListString(newJob.Skills)

		err = db.InsertJob(newJob.DateLine, newJob.Title,
			newJob.Description, resp, skills,
			newJob.Address, newJob.SalaryRangeFrom,
			newJob.SalaryRangeTo, newJob.EmployementType, user.AccountId, cc.Db())
		if err != nil {
			data["errorMsgs"] = []string{"Something went wrong. Please try again."}
			return renderWithAuthContext("create-job.html", c, data)
		}
		data["msg"] = "Job post created"
	}

	return renderWithAuthContext("create-job.html", c, data)
}

// Delete owned jobs or posted where no one has applied
func DeleteJob(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	user, err := GetUserFromSession(c, cc.Db())
	var job struct {
		ID int `form:"job_id" validate:"required"`
	}

	if err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusSeeOther, "/users/profile?info=posted")
	}

	if err = cc.Bind(&job); err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusSeeOther, "/users/profile?info=posted")
	}

	if err = cc.Validate(job); err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusSeeOther, "/users/profile?info=posted")
	}

	owned, err := db.IsJobOwned(job.ID, user.AccountId, cc.Db())
	if err != nil {
		log.Println(err)
	} else {
		if owned {
			err := db.RemoveJob(job.ID, user.AccountId, cc.Db())
			if err != nil {
				log.Println(err)
			}
		}
	}

	return cc.Redirect(http.StatusSeeOther, "/users/profile?info=posted")
}

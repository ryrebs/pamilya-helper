package routes

import (
	"log"
	"net/http"
	"pamilyahelper/webapp/server/db"
	"strings"

	"github.com/labstack/echo/v4"
)

// Verify account view
func VerifyAccountView(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	data := map[string]interface{}{
		"is_log_in": true,
	}
	user, err := GetUserFromSession(cc, cc.Db())
	if user != nil {

		// Redirect if already verified
		if user.IsVerified {
			return c.Redirect(http.StatusSeeOther, "/users/profile")
		}
		data["data"] = map[string]interface{}{
			"name":         user.Name,
			"email":        user.Email,
			"birthdate":    user.Birthdate,
			"address":      user.Address,
			"is_admin":     user.IsAdmin,
			"is_verified":  user.IsVerified,
			"gov_id_image": user.GovIDImage,
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
	file, err := cc.FormFile("gov_id_image")
	if err != nil {
		log.Println(err.Error())
		errorMsgs = append(errorMsgs, "Invalid File")
		data["msgs"] = errorMsgs
		return renderWithAuthContext("verify-profile.html", cc, data)
	}

	// Update user details
	err = db.UpdateUserDetail(*user, editableUser, file, cc.Db())
	if err != nil {
		data["msgs"] = []string{"Something went wrong. Please try again later."}
		return renderWithAuthContext("verify-profile.html", cc, data)
	}
	data["success_msg"] = "Success. Waiting for approval."
	data["data"] = map[string]interface{}{
		"name":         editableUser.Name,
		"email":        user.Email,
		"birthdate":    editableUser.Birthdate,
		"address":      editableUser.Address,
		"is_admin":     user.IsAdmin,
		"is_verified":  user.IsVerified,
		"gov_id_image": file.Filename,
	}
	return renderWithAuthContext("verify-profile.html", cc, data)
}

func VerifyUser(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	user := struct {
		Action string `form:"action" validate:"oneof=accept reject"`
		Email  string `form:"email" validate:"required"`
	}{}
	if err := cc.Bind(&user); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if err := cc.Validate(user); err != nil {
		log.Println(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	is_verify := 0
	if user.Action == "accept" {
		is_verify = 1
	}
	db.UpdateUserVerification(user.Email, is_verify, cc.Db())

	return cc.Redirect(http.StatusSeeOther, "/users/profile")
}

func UploadProfileImage(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	user, err := GetUserFromSession(cc, cc.Db())

	if err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusOK, "/users/profile")
	}

	file, err := cc.FormFile("profile_image")
	if err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusSeeOther, "/users/profile")
	}

	// Update user details
	err = db.UpdateUserDetailProfileImage(user.AccountId, file, cc.Db())
	if err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusSeeOther, "/users/profile")
	}
	return cc.Redirect(http.StatusSeeOther, "/users/profile")
}

func UploadITRFile(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	user, err := GetUserFromSession(cc, cc.Db())

	if err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusOK, "/users/profile")
	}

	file, err := cc.FormFile("income_tax_return")
	if err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusSeeOther, "/users/profile")
	}

	// Update user details
	err = db.UpdateUserDetailITRFile(user.AccountId, file, cc.Db())
	if err != nil {
		log.Println(err)
		return cc.Redirect(http.StatusSeeOther, "/users/profile")
	}
	return cc.Redirect(http.StatusSeeOther, "/users/profile")
}

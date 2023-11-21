package routes

import (
	"fmt"
	"net/http"
	"pamilyahelper/webapp/server/db"

	"github.com/labstack/echo/v4"
)

func SignIn(c echo.Context) error {
	cc := c.(*db.CustomDBContext)
	r := cc.Request()

	if r.Method == "GET" {
		return cc.Render(http.StatusOK, "signin-signup.html", nil)
	}

	if r.Method == "POST" {
		name := cc.FormValue("name")
		password := cc.FormValue("password")

		fmt.Print(name, password)
		return cc.Render(http.StatusOK, "signin-signup.html", nil)
	}

	return cc.Render(http.StatusMethodNotAllowed, "signin-signup.html", nil)
}

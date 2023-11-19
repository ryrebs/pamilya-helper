package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SignIn(c echo.Context) error {
	return c.Render(http.StatusOK, "signin-signup.html", nil)
}

package routes

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
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
	return renderWithAuthContext("profile.html", c, map[string]interface{}{
		"is_log_in": true,
	})
}

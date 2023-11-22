package routes

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func Index(c echo.Context) error {
	sess, _ := session.Get("auth-pamilyahelper-session", c)

	if sess != nil && sess.Values["user"] != nil {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"is_log_in": true,
		})
	}

	return c.Render(http.StatusOK, "index.html", nil)
}

func About(c echo.Context) error {
	return c.Render(http.StatusOK, "about.html", nil)
}

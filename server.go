package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"

	"pamilyahelper/webapp/routes"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Static("static", "public/static")
	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("public/templates/*.html")),
	}
	e.GET("/", routes.Index)
	e.GET("/about", routes.About)

	users := e.Group("users")
	users.Add("GET", "/signin", routes.SignIn)

	e.Logger.Fatal(e.Start("127.0.0.1:5000"))
}

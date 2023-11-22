package server

import (
	"database/sql"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/labstack/echo/v4"

	"pamilyahelper/webapp/server/db"
	"pamilyahelper/webapp/server/routes"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func initDBiFNotExists() *sql.DB {
	if _, err := os.Stat(db.DefaultPamilyaHelperDBName); err != nil {
		log.Printf("%v database not found. Creating db...", db.DefaultPamilyaHelperDBName)
		return db.InitDB()
	}
	return db.GetDBConn(db.DefaultPamilyaHelperDBName)
}

func Serve() {
	dbConn := initDBiFNotExists()
	if dbConn != nil {
		log.Println("Starting server...")

		defer dbConn.Close()

		// Init echo app
		e := echo.New()
		e.Use(session.Middleware(sessions.NewCookieStore([]byte("session-key-replace-me-in-prod"))))

		// Setup static files and templates
		e.Static("static", "public/static")
		e.Renderer = &Template{
			templates: template.Must(template.ParseGlob("public/templates/*.html")),
		}

		// Custom Middlewares
		e.Use(db.AddDBContextMiddleware(dbConn))

		// Routes
		e.GET("/", routes.Index)
		e.GET("/about", routes.About)

		users := e.Group("users")
		users.Match([]string{"GET", "POST"}, "/signin", routes.SignIn)

		e.Logger.Fatal(e.Start("127.0.0.1:5000"))
	} else {
		log.Println("Unable to start server. Make sure database is initialized.")
	}
}

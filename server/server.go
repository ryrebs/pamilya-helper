package server

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

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
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
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
		e.Match([]string{"GET", "POST"}, "/signin", routes.SignIn, routes.RedirectToProfileMiddleware)
		e.POST("/signup", routes.SignUp)

		// Routes - authenticated users
		users := e.Group("users", routes.RedirectToSignInMiddleware)
		users.Match([]string{"GET", "PATCH"}, "/profile", routes.Profile)
		users.POST("/signout", routes.SignOut)

		// Util routes - for dev and priviledged users
		// NOTE: Don't expose or serve on prod.
		unprotected := e.Group("unprotected")
		unprotected.DELETE("/user", routes.RemoveUser)

		default_port := "5000"
		if port, exist := os.LookupEnv("PORT"); exist {
			default_port = port
		}
		default_local_ip := "127.0.0.1"
		if local_ip, exist := os.LookupEnv("EXPOSE_IP"); exist {
			default_local_ip = local_ip
		}

		e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", default_local_ip, default_port)))
	} else {
		log.Println("Unable to start server. Make sure database is initialized.")
	}
}

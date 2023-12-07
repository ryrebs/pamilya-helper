package server

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"pamilyahelper/webapp/server/db"
	"pamilyahelper/webapp/server/routes"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
)

var RequestLimit = 100

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

func createUploadFolder() error {
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); err != nil {
		log.Printf("Creating uploads folder...")
		err := os.Mkdir(uploadDir, 0755)
		if err != nil {
			log.Println(err.Error())
		}
		return err
	}
	return nil
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func Serve() {
	// Initialize necessary components.
	err := createUploadFolder()
	dbConn := initDBiFNotExists()

	// Start the server
	if dbConn != nil && err == nil {
		log.Println("Starting server...")
		defer dbConn.Close()

		// Init echo app
		e := echo.New()
		e.Validator = &CustomValidator{validator: validator.New()}
		// e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.Use(session.Middleware(sessions.NewCookieStore([]byte("session-key-replace-me-in-prod"))))

		// Setup static files and templates
		e.Static("static", "public/static")   // html, js, css
		e.Static("uploads", "public/uploads") // serve uploaded files

		// Parse templates and create custom template functions
		e.Renderer = &Template{
			templates: template.Must(template.ParseGlob("public/templates/*.html")),
		}

		// Custom Middlewares
		e.Use(db.AddDBContextMiddleware(dbConn))

		// Routes
		e.GET("/", routes.Index)
		e.GET("/about", routes.About)
		e.GET("/jobs", routes.JobList)
		e.GET("/helpers", routes.Helper)
		e.GET("/jobs/view/:id", routes.JobDetail)
		e.POST("/signup", routes.SignUp)
		e.GET("/contact", routes.Contact)
		e.POST("/upload/profileimage", routes.UploadProfileImage, routes.RequireSignInMiddleware)
		e.Match([]string{"GET", "POST"}, "/signin", routes.SignIn, routes.RedirectToProfileMiddleware)

		// Routes - for authenticated admin
		admin := e.Group("admin", routes.RequireSignInMiddleware, routes.RequireAdminMiddleware)
		admin.POST("/verify/user", routes.VerifyUser)

		// Routes - authenticated users
		users := e.Group("users", middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(RequestLimit))), routes.RequireSignInMiddleware)
		users.Match([]string{"GET", "POST"}, "/profile", routes.Profile)
		users.POST("/signout", routes.SignOut)
		users.Match([]string{"GET", "POST"}, "/profile/verify", routes.VerifyAccountView)

		jobs := e.Group("jobs",
			middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(RequestLimit))),
			routes.RequireSignInMiddleware,
			routes.RequireVerifiedUserMiddleware)
		jobs.Match([]string{"GET", "POST"}, "/create", routes.CreateJob)
		jobs.POST("/delete", routes.DeleteJob)
		jobs.POST("/view/:id", routes.JobDetail)

		// Util routes - for dev or privileged access
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

package server

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
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

		// Load fixtures
		db.LoadFixtures()

		// Init echo app
		e := echo.New()
		e.Validator = &CustomValidator{validator: validator.New()}
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.Use(session.Middleware(sessions.NewCookieStore([]byte("session-key-replace-me-in-prod"))))

		// Setup static files and templates
		e.Static("static", "public/static")   // html, js, css
		e.Static("uploads", "public/uploads") // serve uploaded files
		e.Renderer = &Template{
			templates: template.Must(template.ParseGlob("public/templates/*.html")),
		}

		// Custom Middlewares
		e.Use(db.AddDBContextMiddleware(dbConn))

		// Check session before accessing uploads/
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				targetURI := c.Request().RequestURI
				if strings.Contains(targetURI, "/uploads") {
					sess, _ := session.Get("auth-pamilyahelper-session", c)
					cc := c.(*db.CustomDBContext)
					err := routes.CheckSessionExist(c, sess, cc.Db())
					if err != nil {
						return c.Redirect(http.StatusSeeOther, "/signin")
					}
				}
				return next(c)
			}
		})

		// Routes
		e.GET("/", routes.Index)
		e.GET("/about", routes.About)
		e.Match([]string{"GET", "POST"}, "/signin", routes.SignIn, routes.RedirectToProfileMiddleware)
		e.POST("/signup", routes.SignUp)

		// Routes - for authenticated admin
		admin := e.Group("admin", routes.RequireSignInMiddleware, routes.RequireAdminMiddleware)
		admin.POST("/verify/user", routes.VerifyUser)

		// Routes - authenticated users
		users := e.Group("users", routes.RequireSignInMiddleware)
		users.GET("/profile", routes.Profile)
		users.POST("/signout", routes.SignOut)
		users.Match([]string{"GET", "POST"}, "/profile/verify", routes.VerifyAccountView)

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

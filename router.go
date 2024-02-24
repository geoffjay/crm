package main

import (
	"net/http"
	"time"

	"github.com/geoffjay/crm/handlers"
	"github.com/geoffjay/crm/util"
	"github.com/geoffjay/crm/views"
	"github.com/geoffjay/crm/views/pages"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	log "github.com/sirupsen/logrus"
)

const (
	Development = "development"
)

func csrfErrorHandler(c *fiber.Ctx, err error) error {
	// Log the error so we can track who is trying to perform CSRF attacks
	// customize this to your needs
	log.WithFields(log.Fields{
		"service": "app",
		"context": "router.csrfErrorHandler",
		"error":   err,
		"ip":      c.IP(),
		"request": c.OriginalURL(),
	}).Error("CSRF Error")

	log.Debugf("ctx: %v", c)

	// check accepted content types
	switch c.Accepts("html", "json") {
	case "json":
		// Return a 403 Forbidden response for JSON requests
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "403 Forbidden",
		})
	case "html":
		c.Locals("title", "Error")
		c.Locals("error", "403 Forbidden")
		c.Locals("errorCode", "403")

		// Return a 403 Forbidden response for HTML requests
		return views.Render(c, pages.Error(), templ.WithStatus(http.StatusForbidden))
	default:
		// Return a 403 Forbidden response for all other requests
		return c.Status(fiber.StatusForbidden).SendString("403 Forbidden")
	}
}

func httpHandler(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(f)
}

func initRouter(app *fiber.App) {
	staticContents := util.Getenv("CRM_PUBLIC_PATH", "./public")

	csrfConfig := csrf.Config{
		Session:        handlers.SessionStore,
		KeyLookup:      "form:csrf",
		CookieName:     "__Host-csrf",
		CookieSameSite: "Lax",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		ContextKey:     "csrf",
		ErrorHandler:   csrfErrorHandler,
		Expiration:     30 * time.Minute,
	}
	csrfMiddleware := csrf.New(csrfConfig)

	app.Static("/public", staticContents)

	app.Get("/", handlers.IndexPage)
	// app.Get("/login", csrfMiddleware, handlers.LoginPage)
	// app.Post("/login", csrfMiddleware, handlers.Login)
	// app.Get("/logout", handlers.Logout)
	// app.Post("/register", handlers.Register)

	admin := app.Group("/admin")
	admin.Use(csrfMiddleware)
	admin.Get("/", handlers.AdminIndexPage)
	admin.Get("/pages", handlers.AdminPagesPage)
	admin.Get("/pages/:id", handlers.AdminPagePage)
	admin.Get("/pages/:id/sections", handlers.AdminSectionsPage)
	admin.Get("/pages/:id/sections/:section", handlers.AdminSectionPage)

	// API routes
	api := app.Group("/api")
	v1 := api.Group("/v1", func(c *fiber.Ctx) error {
		c.Set("Version", "v1")
		return c.Next()
	})

	// TODO: create auth middleware that checks for token
	//
	// v1.Use(authMiddleware)
	v1.Get("/pages", handlers.GetPages)
	v1.Get("/pages/:id", handlers.GetPage)
	v1.Get("/pages/:id/sections", handlers.GetSections)
	v1.Get("/pages/:id/sections/:section", handlers.GetSection)

	// Development routes
	// env := util.Getenv("APP_ENV", "development")
	// if strings.ToLower(env) == Development {
	// 	log.Debug("Development routes enabled")
	//
	// 	dev := app.Group("/dev")
	// 	dev.Get("/reload", adaptor.HTTPHandler(httpHandler(handlers.Reload)))
	// }

	app.Use(handlers.NotFound)
}

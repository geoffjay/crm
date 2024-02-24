package handlers

import (
	"net/http"

	"github.com/geoffjay/crm/views"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// SessionStore app wide session store.
var SessionStore *session.Store

// Index renders the application index page.
func Index(c *fiber.Ctx) error {
	session, err := SessionStore.Get(c)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	loggedIn, _ := session.Get("loggedIn").(bool)
	if !loggedIn {
		return c.Redirect("/login")
	}

	c.Locals("title", "App")

	return views.Render(c, views.Index(), templ.WithStatus(http.StatusOK))
}

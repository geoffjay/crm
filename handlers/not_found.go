package handlers

import (
	"net/http"

	"github.com/geoffjay/crm/views"
	"github.com/geoffjay/crm/views/pages"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

func NotFound(c *fiber.Ctx) error {
	return views.Render(c, pages.NotFound(), templ.WithStatus(http.StatusNotFound))
}

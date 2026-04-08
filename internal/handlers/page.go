package handlers

import "github.com/gofiber/fiber/v3"

// IndexPage отдаёт главную HTML-страницу приложения.
func IndexPage(c fiber.Ctx) error {
	return c.SendFile("./web/index.html")
}

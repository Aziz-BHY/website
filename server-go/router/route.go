package router

import (
	"website/server/controller"
	"website/server/render"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Events
	app.Post("/events", controller.CreateEvent)
	app.Get("/events", controller.GetEvents)
	app.Put("/events/attendee", controller.AddAttendee)
	app.Put("/events/:id", controller.UpdateEvent)
	app.Delete("/events/:id", controller.DeleteEvent)
	//newsletter
	app.Post("/newsletter", controller.AddEmail)
	app.Get("/newsletter", controller.GetCSV)
	//templates
	app.Get("/admin", render.Index)
	app.Get("/admin/events", render.Events)
	app.Get("/admin/events/create", render.CreateEvent)
	app.Get("/admin/events/:id", render.UpdateEvent)
	app.Get("/admin/transactions", render.GetTransactions)
}

package main

import (
	"log"
	"website/server/config"
	"website/server/database"
	"website/server/router"

	"github.com/gofiber/template/html"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {

	config.GetSVC()

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	router.SetupRoutes(app)
	app.Listen(":5000")
}

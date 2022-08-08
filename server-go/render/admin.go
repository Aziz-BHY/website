package render

import (
	"context"
	"os"
	"website/server/database"
	"website/server/models"
	"website/server/resources"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func Index(c *fiber.Ctx) error {
	//read file
	url, err := os.ReadFile("variables/SERVER_EXTERNAL_URL")
	if err != nil {
		return err
	}
	return c.Render("index", fiber.Map{
		"URL": string(url),
	})
}

func Events(c *fiber.Ctx) error {
	//read file
	url, err := os.ReadFile("variables/SERVER_EXTERNAL_URL")
	if err != nil {
		return err
	}
	events, err := resources.GetEvents(c.Query("q"))
	if err != nil {
		return c.Render("error", fiber.Map{})
	}
	return c.Render("getEvents", fiber.Map{
		"URL":    string(url),
		"Events": events.Events,
	})
}

func CreateEvent(c *fiber.Ctx) error {
	//read file
	url, err := os.ReadFile("variables/SERVER_EXTERNAL_URL")
	if err != nil {
		return err
	}
	return c.Render("createEvent", fiber.Map{
		"URL": string(url),
	})
}

func UpdateEvent(c *fiber.Ctx) error {
	//read file
	url, err := os.ReadFile("variables/SERVER_EXTERNAL_URL")
	if err != nil {
		return err
	}
	event, err := resources.GetEventById(c.Params("id"))
	if err != nil {
		return c.Render("error", fiber.Map{})
	}
	return c.Render("updateEvent", fiber.Map{
		"URL":   string(url),
		"Event": event,
	})
}

func GetTransactions(c *fiber.Ctx) error {
	transactions := []models.Transaction{}

	cursor, err := database.TransColl.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Render("error", fiber.Map{})
	}
	if err := cursor.Err(); err != nil {
		return c.Render("error", fiber.Map{})
	}
	if err = cursor.All(context.TODO(), &transactions); err != nil {
		return c.Render("error", fiber.Map{})
	}
	cursor.Close(context.TODO())
	return c.Render("transactions", fiber.Map{
		"Transactions": transactions,
	})
}

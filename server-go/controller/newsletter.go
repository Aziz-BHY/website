package controller

import (
	"context"
	"net/mail"
	"website/server/database"
	m "website/server/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func AddEmail(c *fiber.Ctx) error {
	sub := new(m.Subscription)

	if err := c.BodyParser(sub); err != nil {
		return err
	}

	if _, err := mail.ParseAddress(sub.Email); err != nil {
		return c.JSON("email field incorrect")
	}

	tempSub := m.Subscription{}

	err := database.NewsletterColl.FindOne(context.TODO(), bson.M{"email": sub.Email}).Decode(&tempSub)
	if err != nil {
		_, err := database.NewsletterColl.InsertOne(context.Background(), sub)
		if err != nil {
			return err
		}
	}

	return c.JSON("inserted")

}

func GetCSV(c *fiber.Ctx) error {
	subs := m.Subscriptions{}

	cursor, err := database.NewsletterColl.Find(context.TODO(), bson.M{})
	if err != nil {
		return err
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if err = cursor.All(context.TODO(), &subs.Subs); err != nil {
		return err
	}
	cursor.Close(context.TODO())
	out := "email,name \n"

	for _, sub := range subs.Subs {

		out = out + sub.Email + "," + sub.Name + "\n"
	}

	c.Attachment("subs.csv")
	return c.SendString(out)
}

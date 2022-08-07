package controller

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"
	"website/server/config"
	"website/server/database"
	"website/server/models"
	"website/server/resources"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/calendar/v3"
)

func GetEvents(c *fiber.Ctx) error {
	events, err := resources.GetEvents(c.Query("q"))
	if err != nil {
		return err
	}
	return c.JSON(events.Events)
}

func DeleteEvent(c *fiber.Ctx) error {
	// convert id string to ObjectId
	objectId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return err
	}
	_, err = database.EventColl.DeleteOne(context.TODO(), bson.M{"_id": objectId})
	if err != nil {
		return err
	}
	return c.JSON("deleted")
}

func UpdateEvent(c *fiber.Ctx) error {
	//parse event
	event := new(models.Event)
	if err := c.BodyParser(event); err != nil {
		return err
	}
	// convert id string to ObjectId
	objectId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return err
	}
	//update event (mongo)
	_, err = database.EventColl.UpdateByID(context.TODO(), objectId, bson.M{
		"$set": models.Event{Price: event.Price, ImageLink: event.ImageLink},
	})
	if err != nil {
		return err
	}

	//update event (calendar)
	_, err = config.Srv.Events.Patch("primary", event.CalendarId, &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime,
			TimeZone: "Africa/Tunis",
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime,
			TimeZone: "Africa/Tunis",
		},
	}).Do()

	return c.JSON("updated")

}

func AddAttendee(c *fiber.Ctx) error {
	//Parse Body
	attendee := new(models.Attendee)
	if err := c.BodyParser(attendee); err != nil {
		return err
	}

	//Get Event Info
	event, err := config.Srv.Events.Get("primary", attendee.EventId).Do()
	if err != nil {
		return err
	}
	tempEvent := models.Event{}

	err = database.EventColl.FindOne(context.TODO(), bson.M{"calendarId": attendee.EventId}).Decode(&tempEvent)
	//Check if evenement is free
	if tempEvent.Price > 0 {
		if attendee.PaymentToken == "" {
			return errors.New("this event is not free, need payment token")
		}
		//get values from files
		paymeeURL, err := os.ReadFile("variables/PAYMEE_URL")
		if err != nil {
			return err
		}
		paymeekey, err := os.ReadFile("secrets/PAYMEE_API_KEY")
		if err != nil {
			return err
		}

		//send Get request to Paymee
		client := &http.Client{}
		req, err := http.NewRequest("GET", string(paymeeURL)+"/api/v1/payments/"+attendee.PaymentToken+"/check", nil)
		req.Header.Set("Authorization", "Token "+string(paymeekey))
		resp, err := client.Do(req)
		defer resp.Body.Close()
		if err != nil {
			return err
		}
		//Parse paymee response
		var paymeeResp models.PaymeeResponse

		if err := json.NewDecoder(resp.Body).Decode(&paymeeResp); err != nil {
			log.Fatal(err)
		}
		//check if amount is same
		if paymeeResp.Data.Amount != tempEvent.Price {
			return errors.New("Amount is not same, error ")
		}

		//save transaction
		_, err = database.TransColl.InsertOne(context.Background(),
			models.Transaction{
				CalendarId:    tempEvent.CalendarId,
				Token:         paymeeResp.Data.Token,
				BuyerId:       paymeeResp.Data.BuyerId,
				TransactionID: paymeeResp.Data.TransactionId,
				Amount:        paymeeResp.Data.Amount,
				Date:          time.Now(),
				Email:         attendee.Email,
			})

		if err != nil {
			log.Fatalln("Insert:", err)
			return err
		}

	}
	//Add new Email
	event.Attendees = append(event.Attendees,
		&calendar.EventAttendee{Email: attendee.Email},
	)

	//Update event
	_, err = config.Srv.Events.Patch("primary", attendee.EventId, event).Do()
	if err != nil {
		return err
	}
	return c.JSON("email added")
}

func CreateEvent(c *fiber.Ctx) error {

	if err := deletePastEvents(); err != nil {
		return err
	}
	//Parse Body
	event := new(models.Event)
	if err := c.BodyParser(event); err != nil {
		return err
	}

	//Fill Event Info
	NewEvent := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime,
			TimeZone: "Africa/Tunis",
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime,
			TimeZone: "Africa/Tunis",
		},
		Attendees:               []*calendar.EventAttendee{},
		GuestsCanInviteOthers:   &[]bool{false}[0],
		GuestsCanSeeOtherGuests: &[]bool{false}[0],
		ConferenceData: &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId: uuid.NewString(),
				ConferenceSolutionKey: &calendar.ConferenceSolutionKey{
					Type: "hangoutsMeet",
				},
			},
		},
	}

	//Create new Calendar Event
	calendarId := "primary"
	NewEvent, err := config.Srv.Events.Insert(calendarId, NewEvent).ConferenceDataVersion(1).Do()
	if err != nil {
		return err
	}

	//insert into DB
	insertResult, err := database.EventColl.InsertOne(context.Background(), models.Event{CalendarId: NewEvent.Id, Price: event.Price, ImageLink: event.ImageLink})
	if err != nil {
		log.Fatalln("Insert:", err)
		return err
	}
	return c.JSON(insertResult)
}

func deletePastEvents() error {
	myEvents := models.Events{}

	//get all events from DB
	cursor, err := database.EventColl.Find(context.TODO(), bson.M{})
	if err != nil {
		return err
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if err = cursor.All(context.TODO(), &myEvents.Events); err != nil {
		return err
	}
	cursor.Close(context.TODO())

	for index := range myEvents.Events {
		//Get Event Info from Google API
		event, err := config.Srv.Events.Get("primary", myEvents.Events[index].CalendarId).Do()
		if err != nil {
			return err
		}
		// convert string to time
		date, err := time.Parse("2006-01-02", event.Start.DateTime[:10])

		if err != nil {
			return err
		}
		// if event occurred
		if date.Before(time.Now()) {
			_, err := database.EventColl.DeleteOne(context.TODO(), bson.M{"_id": myEvents.Events[index].Id})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

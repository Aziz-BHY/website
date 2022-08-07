package resources

import (
	"context"
	"strings"
	"website/server/config"
	"website/server/database"
	"website/server/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetEvents(q string) (models.Events, error) {
	myEvents := models.Events{}
	responseEvents := models.Events{}

	cursor, err := database.EventColl.Find(context.TODO(), bson.M{})
	if err != nil {
		return responseEvents, err
	}
	if err := cursor.Err(); err != nil {
		return responseEvents, err
	}
	if err = cursor.All(context.TODO(), &myEvents.Events); err != nil {
		return responseEvents, err
	}
	cursor.Close(context.TODO())

	for index := range myEvents.Events {
		//Get Event Info from Google API
		event, err := config.Srv.Events.Get("primary", myEvents.Events[index].CalendarId).Do()
		if err != nil {
			return responseEvents, err
		}
		if q != "" {
			if strings.Contains(event.Summary, q) || strings.Contains(event.Description, q) {
				myEvents.Events[index].Summary = event.Summary
				myEvents.Events[index].Description = event.Description
				myEvents.Events[index].StartTime = event.Start.DateTime
				myEvents.Events[index].EndTime = event.End.DateTime
				responseEvents.AddItem(myEvents.Events[index])
			}
		} else {
			myEvents.Events[index].Summary = event.Summary
			myEvents.Events[index].Description = event.Description
			myEvents.Events[index].StartTime = event.Start.DateTime
			myEvents.Events[index].EndTime = event.End.DateTime
			responseEvents.AddItem(myEvents.Events[index])
		}

	}
	return responseEvents, nil
}

func GetEventById(id string) (models.Event, error) {
	event := models.Event{}
	// convert id string to ObjectId
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return event, nil
	}

	err = database.EventColl.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&event)
	if err != nil {
		return event, nil
	}
	calendarEvent, err := config.Srv.Events.Get("primary", event.CalendarId).Do()
	if err != nil {
		return event, nil
	}
	event.Summary = calendarEvent.Summary
	event.Description = calendarEvent.Description
	event.StartTime = calendarEvent.Start.DateTime
	event.EndTime = calendarEvent.End.DateTime
	return event, nil
}

package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"website/events/models"

	"github.com/gofiber/fiber/v2"
	clov "github.com/ostafen/clover"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getSVC() *calendar.Service {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}
	return srv
}

func GetEvents(c *fiber.Ctx) error {
	//Get Service
	srv := getSVC()

	//Open Db
	db, _ := clov.Open("events-db")
	defer db.Close()

	docs, _ := db.Query("Events").FindAll()

	myEvents := []models.Event{}
	tempEvent := new(models.EventDB)

	for _, doc := range docs {
		doc.Unmarshal(tempEvent)
		//Get Event Info from Google API
		event, err := srv.Events.Get("primary", tempEvent.EventId).Do()
		if err != nil {
			return err
		}
		myEvents = append(myEvents, models.Event{
			Id:          event.Id,
			Summary:     event.Summary,
			Description: event.Description,
			StartTime:   event.Start.DateTime,
			EndTime:     event.End.DateTime,
			Price:       tempEvent.Price,
		})
	}
	return c.JSON(myEvents)

}

func AddAttendee(c *fiber.Ctx) error {
	//Get Service
	srv := getSVC()

	//Parse Body
	attendee := new(models.Attendee)
	if err := c.BodyParser(attendee); err != nil {
		return err
	}

	//Get Event Info
	event, err := srv.Events.Get("primary", attendee.EventId).Do()
	if err != nil {
		return err
	}

	//Open DB
	db, _ := clov.Open("events-db")
	defer db.Close()
	db.CreateCollection("Events")

	//search for event by eventId
	doc, _ := db.Query("Events").Where(clov.Field("eventId").Eq(attendee.EventId)).FindFirst()

	//parse doc to tempEvent
	tempEvent := new(models.EventDB)
	doc.Unmarshal(tempEvent)

	if tempEvent.Price > 0 {
		if attendee.PaymentToken == "" {
			return errors.New("this event is not free, need payment token")
		}
		//send Get request to Paymee
		client := &http.Client{}
		req, err := http.NewRequest("GET", os.Getenv("PAYMEE_URL")+"/api/v1/payments/"+attendee.PaymentToken+"/check", nil)
		req.Header.Set("Authorization", "Token "+os.Getenv("PAYMEE_API_KEY"))
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
	}
	//Add new Email
	event.Attendees = append(event.Attendees,
		&calendar.EventAttendee{Email: attendee.Email},
	)

	//Update event
	_, err = srv.Events.Patch("primary", attendee.EventId, event).Do()
	if err != nil {
		return err
	}
	return c.JSON("email added")
}

func CreateEvent(c *fiber.Ctx) error {
	//Get Service
	srv := getSVC()

	//Open DB
	db, _ := clov.Open("events-db")
	defer db.Close()
	db.CreateCollection("Events")

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
	}

	//Create new Calendar Event
	calendarId := "primary"
	NewEvent, err := srv.Events.Insert(calendarId, NewEvent).Do()
	if err != nil {
		return err
	}

	//Add event to DB
	doc := clov.NewDocument()
	doc.Set("eventId", NewEvent.Id)
	doc.Set("price", event.Price)

	_, err = db.InsertOne("Events", doc)
	if err != nil {
		return err
	}

	return c.JSON("inserted")
}
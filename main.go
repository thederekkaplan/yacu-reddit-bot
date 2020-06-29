package main

import (
	"net/http"
	"os"
	"fmt"
	"log"
	"context"
	"encoding/json"
	"time"
	"strconv"
	"io/ioutil"

	"github.com/turnage/graw/reddit"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type MobilizeEvent struct {
	Title string
	Description string
	Url string `json:"browser_url"`
	Timeslots []Timeslot
}

type Timeslot struct {
	Id float64
	StartDate float64 `json:"start_date"`
	EndDate float64 `json:"end_date"`
}

type Event struct {
	Id float64
	Title string
	Description string
	Url string
	StartDate time.Time
	EndDate time.Time
}

func main() {
	http.HandleFunc("/update", update)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Listening on port", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func postEvents(account reddit.Account, events []Event) {
	for _, event := range(events) {
		log.Println(event.Url)
		account.PostLink(
			"/r/YACUHQ",
			"[" + event.StartDate.Format("Jan 2, 3:04 PM") + " Eastern] " + event.Title,
			event.Url,
		)
	}
}

func update(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Appengine-Cron") != "true" {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "403 Forbidden")
		return
	}

	cfg := reddit.BotConfig {
		Agent: "golang:github.com/thederekkaplan/yacu-reddit-bot:v1.0.0",
		App: app(),
		Rate: 0,
	}

	bot, err := reddit.NewBot(cfg)
	if err != nil {
		panic(err)
	}

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	now := time.Now().In(loc)
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	start := date.AddDate(0, 0, 1)
	end := date.AddDate(0, 0, 2)

	log.Println("Start time", start.Unix())
	log.Println("End time", end.Unix())

	url := "https://api.mobilize.us/v1/organizations/2596/events?timeslot_start=gte_" +
		strconv.Itoa(int(start.Unix())) + 
		"&timeslot_start=lt_" +
		strconv.Itoa(int(end.Unix()))

	events := getEvents(url)

	postEvents(bot, events)
}

func app() reddit.App {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	accessReq := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/commanding-way-273100/secrets/reddit/versions/latest",
	}

	result, err := client.AccessSecretVersion(ctx, accessReq)
	if err != nil {
		panic(err)
	}

	var data reddit.App

	err = json.Unmarshal(result.Payload.Data, &data)
	if err != nil {
		panic(err)
	}

	return data
}

func getEvents(url string) []Event {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var objmap map[string]json.RawMessage
	err = json.Unmarshal(body, &objmap)
	if err != nil {
		panic(err)
	}

	var mobilizeEvents []MobilizeEvent

	err = json.Unmarshal(objmap["data"], &mobilizeEvents)
	if err != nil {
		panic(err)
	}

	var events []Event

	// Convert timeslots into separate events
	for _, event := range mobilizeEvents {
		for _, timeslot := range event.Timeslots {
			events = append(events, Event{
				Id: timeslot.Id,
				Title: event.Title,
				Description: event.Description,
				Url: event.Url,
				StartDate: time.Unix(int64(timeslot.StartDate), 0),
				EndDate: time.Unix(int64(timeslot.EndDate), 0),
			})
		}
	}

	return events
}

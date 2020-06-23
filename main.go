package main

import (
	"net/http"
	"os"
	"fmt"
	"context"

	"github.com/turnage/graw/reddit"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

func main() {
	http.HandleFunc("/update", update)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Listening on port", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func update(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Appengine-Cron") != true {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "403 Forbidden")
		return
	}

	cfg := reddit.BotConfig {
		Agent: "golang:github.com/thederekkaplan/yacu-reddit-bot:v1.0.0"
		App: app()
	}

}

func app() reddit.App {
	projectID := "commanding-way-273100"

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	
}
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"spaghetti/pkg/formatmessage"
	"spaghetti/pkg/postmessage"
	"time"

	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func now() string {
	return time.Now().Format(time.RFC3339)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	slackAPI := slack.New(os.Getenv("SLACK_TOKEN"))
	channelID := os.Getenv("SLACK_CHANNEL_ID")

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, req *http.Request) {
		var body formatmessage.Webhook
		event := req.Header.Get("X-GitHub-Event")
		deliveryID := req.Header.Get("X-GitHub-Delivery")
		fmt.Printf("[%s] event: %s, deliveryID: %s\n", now(), event, deliveryID)

		if event == "pull_request" {
			err := json.NewDecoder(req.Body).Decode(&body)
			if err != nil {
				log.Fatalf("json decode: %s", err)
			}

			fmt.Printf("[%s]: %s\n", now(), body.Action)

			if body.Action == "review_requested" && (body.RequestedTeam.Name != "" || len(body.PullRequest.RequestedTeams) > 0) {
				s, _ := json.MarshalIndent(body, "", "\t")
				fmt.Printf("[%s]: %s", now(), string(s))

				message := formatmessage.FormatMessage(body)

				options := postmessage.PostMessageOptions{
					Message:   message,
					ChannelID: channelID,
				}

				postmessage.PostMessage(slackAPI, options)
			}
		}

	})

	fmt.Printf("[%s] started serving", now())
	err = http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		log.Fatalf("http server: %s", err)
	}
}

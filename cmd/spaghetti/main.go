package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func client(ctx context.Context) {
	// app_id, err := strconv.Atoi(os.Getenv("APP_ID"))
	app_id, err := strconv.ParseInt(os.Getenv("APP_ID"), 10, 64)
	if err != nil {
		log.Fatalf("app_id: %s", err)
	}
	installation_id, err := strconv.ParseInt(os.Getenv("INSTALLATION_ID"), 10, 64)
	if err != nil {
		log.Fatalf("app_id: %s", err)
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, app_id, installation_id, "spaghettiapp.2021-09-16.private-key.pem")
	if err != nil {
		log.Fatalf("key: %s", err)
	}
	client := github.NewClient(&http.Client{Transport: itr})

	pr, _, err := client.PullRequests.Get(ctx, "spaghettigc", "spaghetti", 3)
	fmt.Printf("%v\n", pr.RequestedReviewers)
	fmt.Printf("\n\n")
	bod := *pr.Body
	fmt.Printf("%s", bod)

	if err != nil {
		log.Fatalf("client: %s", err)
	}

}

func now() string {
	return time.Now().Format(time.RFC3339)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	slackAPI := slack.New(os.Getenv("SLACK_TOKEN"))

	groups, err := slackAPI.GetUserGroups()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	for _, group := range groups {
		fmt.Printf("ID: %s, Name: %s\n", group.ID, group.Name)
	}

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, req *http.Request) {
		var body Webhook
		event := req.Header.Get("X-GitHub-Event")
		deliveryID := req.Header.Get("X-GitHub-Delivery")
		fmt.Printf("[%s] event: %s, deliveryID: %s\n", now(), event, deliveryID)

		if event == "pull_request" {
			err := json.NewDecoder(req.Body).Decode(&body)
			if err != nil {
				log.Fatalf("json decode: %s", err)
			}
			// fmt.Printf("%+v", body)
			fmt.Printf("[%s]: %s\n", now(), body.Action)

			// requestteams[] length is > 0
			// requestedteam name != ""
			if body.Action == "review_requested" && (body.RequestedTeam.Name != "" || len(body.PullRequest.RequestedTeams) > 0) {
				s, _ := json.MarshalIndent(body, "", "\t")
				fmt.Printf("[%s]: %s", now(), string(s))
			}
			// body, err := io.ReadAll(req.Body)
		}

	})

	fmt.Printf("[%s] started serving", now())
	err = http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		log.Fatalf("http server: %s", err)
	}
}

type Webhook struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		HTMLURL            string    `json:"html_url"`
		Title              string    `json:"title"`
		Body               string    `json:"body"`
		UpdatedAt          time.Time `json:"updated_at"`
		RequestedReviewers []struct {
			Login string `json:"login"`
		} `json:"requested_reviewers"`
		RequestedTeams []struct {
			Name string `json:"name"`
		} `json:"requested_teams"`
	} `json:"pull_request"`
	RequestedTeam struct {
		Name string `json:"name"`
	} `json:"requested_team"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
	Installation struct { // app installation ID at org level
		ID     int    `json:"id"`
		NodeID string `json:"node_id"`
	} `json:"installation"`
}

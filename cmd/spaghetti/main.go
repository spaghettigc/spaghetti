package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"spaghetti/pkg/formatmessage"
	"spaghetti/pkg/postmessage"
	"spaghetti/pkg/vcr"
	"strconv"
	"time"

	"os"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func now() string {
	return time.Now().Format(time.RFC3339)
}

func getPRRequestBody(ctx context.Context, client github.Client) *http.Response {
	pull, _, err := client.PullRequests.Get(ctx, "spaghettigc", "spaghetti", 4)
	if err != nil {
		log.Fatalf("timeline: %s", err)
	}
	htmlURL := pull.GetHTMLURL()

	resp, err := http.Get(htmlURL) // authenticating with GH how??
	if err != nil {
		log.Fatalf("http get: %s", err)
	}
	defer resp.Body.Close()

	return resp
}

func client(ctx context.Context) (*github.Client, error) {
	appID, err := strconv.ParseInt(os.Getenv("APP_ID"), 10, 64)
	if err != nil {
		return nil, err
	}
	installationID, err := strconv.ParseInt(os.Getenv("INSTALLATION_ID"), 10, 64)
	if err != nil {
		return nil, err
	}
	privateKeyFile := os.Getenv("PRIVATE_KEY_FILE")

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, privateKeyFile)
	if err != nil {
		return nil, err
	}
	return github.NewClient(&http.Client{Transport: itr}), nil
}

func main2() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx := context.Background()

	client(ctx)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	slackAPI := slack.New(os.Getenv("SLACK_TOKEN"))
	channelID := os.Getenv("SLACK_CHANNEL_ID")

	ctx := context.Background()

	c, err := client(ctx)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, req *http.Request) {
		var body formatmessage.Webhook

		_, err = vcr.RequestHandler(req, body, "review-multiple-members"+now())

		if err != nil {
			panic(err)
		}

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
				pull, _, err := c.PullRequests.Get(ctx, body.Organization.Login, body.Repository.Name, body.Number)
				if err != nil {
					log.Fatalf("timeline: %s", err)
				}
				htmlURL := pull.GetHTMLURL()

				resp, err := http.Get(htmlURL) // authenticating with GH how??
				if err != nil {
					log.Fatalf("http get: %s", err)
				}
				defer resp.Body.Close()

				assignees, err := formatmessage.GetAssignedReviewersAndTeam(resp, "5705558574") // retunring zero assignees bug?
				fmt.Printf("number of assignees: %d", len(assignees))
				if err != nil {
					log.Fatalf("GetAssignedReviewersAndTeam: %s", err)
				}
				for _, assignee := range assignees {

					message := formatmessage.FormatMessage(body, assignee)

					options := postmessage.PostMessageOptions{
						Message:   message,
						ChannelID: channelID,
					}

					postmessage.PostMessage(slackAPI, options)
				}
			}
		}

	})

	fmt.Printf("[%s] started serving", now())
	err = http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		log.Fatalf("http server: %s", err)
	}
}

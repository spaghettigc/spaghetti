package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
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

func getEventId(ctx context.Context, body formatmessage.Webhook) (string, string, error) {
	var eventID string
	var htmlURL string
	c, err := client(ctx)
	if err != nil {
		return eventID, htmlURL, err
	}

	perPage := 100
	_, response, err := c.Issues.ListIssueTimeline(ctx, body.Organization.Login, body.Repository.Name, body.Number, &github.ListOptions{Page: 1, PerPage: perPage})
	if err != nil {
		return eventID, htmlURL, err
	}
	// look for the event id
	// query the event api

	// fmt.Printf("timeline: %v, %v", *timeline[len(timeline)-2].Event, *timeline[len(timeline)-2].ID)
	// fmt.Printf("LastPage: %v", response.LastPage)
	// fmt.Printf("NextPage: %v", response.NextPage)
	// fmt.Printf("PrevPage: %v", response.PrevPage)
	something := true

	currentPage := response.LastPage

	var eventIDs []string
	// TODO optimise because we're going through all the timeline pages, could pick out last 2 events instead
	for something == true {
		timeline, response, err := c.Issues.ListIssueTimeline(ctx, body.Organization.Login, body.Repository.Name, body.Number, &github.ListOptions{Page: currentPage, PerPage: perPage})
		if err != nil {
			return eventID, htmlURL, err
		}

		for i := len(timeline) - 1; i >= 0; i-- {
			t := timeline[i]
			// fmt.Printf("*t.Event: %v - *t.CreatedAt: %v -  *t.ID: %v\n", *t.Event, *t.CreatedAt, *t.ID)
			if *t.Event == "review_requested" && *t.CreatedAt == body.UpdatedAt {
				eventID = strconv.FormatInt(*t.ID, 10)
				eventIDs = append(eventIDs, eventID)
				something = false
				fmt.Printf("eventID: %v\n", eventID)

			}

		}
		currentPage = response.PrevPage
		fmt.Printf("pagenumber: %v\n", currentPage)
		if currentPage == 0 {
			something = false
			fmt.Println("NOT FOUND")

		}
	}

	pull, _, err := c.PullRequests.Get(ctx, body.Organization.Login, body.Repository.Name, body.Number)
	if err != nil {
		return eventID, htmlURL, err
	}
	htmlURL = pull.GetHTMLURL()

	sort.Strings(eventIDs)
	return eventIDs[1], htmlURL, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	slackAPI := slack.New(os.Getenv("SLACK_TOKEN"))
	channelID := os.Getenv("SLACK_CHANNEL_ID")

	ctx := context.Background()

	// fmt.Printf("eventID: %v", *eventID)

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
				eventID, htmlURL, err := getEventId(ctx, body) // TODO rename getEventId
				if err != nil {
					log.Fatalf("getEventId: %s", err)
				}
				h := fmt.Sprintf("%s#event-%s", htmlURL, eventID)
				fmt.Printf("%s#event-%s", htmlURL, eventID)

				resp, err := http.Get(h) // authenticating with GH how??
				if err != nil {
					log.Fatalf("http get: %s", err)
				}
				defer resp.Body.Close()

				assignees, err := formatmessage.GetAssignedReviewersAndTeam(resp, eventID) // retunring zero assignees bug?
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

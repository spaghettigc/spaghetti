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

	"github.com/allegro/bigcache"
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/store"
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

func getEventIds(ctx context.Context, body formatmessage.Webhook) ([]string, string, error) {
	var eventIDs []string
	var htmlURL string
	c, err := client(ctx)
	if err != nil {
		return eventIDs, htmlURL, err
	}

	perPage := 100
	_, response, err := c.Issues.ListIssueTimeline(ctx, body.Organization.Login, body.Repository.Name, body.Number, &github.ListOptions{Page: 1, PerPage: perPage})
	if err != nil {
		return eventIDs, htmlURL, err
	}
	something := true

	currentPage := response.LastPage

	// TODO optimise because we're going through all the timeline pages, could pick out last 2 events instead
	for something == true {
		var eventID string
		timeline, response, err := c.Issues.ListIssueTimeline(ctx, body.Organization.Login, body.Repository.Name, body.Number, &github.ListOptions{Page: currentPage, PerPage: perPage})
		if err != nil {
			return eventIDs, htmlURL, err
		}

		for i := len(timeline) - 1; i >= 0; i-- {
			t := timeline[i]

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
		return eventIDs, htmlURL, err
	}
	htmlURL = pull.GetHTMLURL()

	return eventIDs, htmlURL, nil
}

type Assigned struct {
	Reviewer string
	Team     string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	slackAPI := slack.New(os.Getenv("SLACK_TOKEN"))
	channelID := os.Getenv("SLACK_CHANNEL_ID")

	ctx := context.Background()

	bigcacheClient, _ := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	bigcacheStore := store.NewBigcache(bigcacheClient, nil) // No options provided (as second argument
	cacheManager := cache.New(bigcacheStore)

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
				eventIDs, htmlURL, err := getEventIds(ctx, body) // TODO rename getEventId
				if err != nil {
					log.Fatalf("getEventId: %s", err)
				}

				// loop over event IDs
				for _, eventID := range eventIDs {
					// get cache

					value, err := cacheManager.Get(eventID)

					if err != nil && err != bigcache.ErrEntryNotFound {
						panic(err)
					}

					if value != nil {
						fmt.Printf("skipped %s\n", eventID)
						continue
					}

					err = cacheManager.Set(eventID, "not nil", nil)
					if err != nil {
						panic(err)
					}

					h := fmt.Sprintf("%s#event-%s", htmlURL, eventID)
					fmt.Printf("%s#event-%s", htmlURL, eventID)

					assignees, err := formatmessage.GetAssignedReviewersAndTeam(eventID, h) // retunring zero assignees bug?
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
		}

	})

	fmt.Printf("[%s] started serving", now())
	err = http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		log.Fatalf("http server: %s", err)
	}
}

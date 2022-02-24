package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"spaghetti/pkg/message"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type PullRequest struct {
	HTMLURL            string              `json:"html_url"`
	Title              string              `json:"title"`
	Body               string              `json:"body"`
	UpdatedAt          time.Time           `json:"updated_at"`
	RequestedReviewers []RequestedReviewer `json:"requested_reviewers"`
	RequestedTeams     []RequestedTeam     `json:"requested_teams"`
}
type RequestedReviewer struct {
	Login string `json:"login"`
}

type RequestedTeam struct {
	Name string `json:"name"`
}

type Repository struct {
	Name string `json:"name"`
}
type Sender struct {
	Login string `json:"login"`
}

type Organization struct {
	Login string `json:"login"`
}

type Webhook struct {
	Action        string `json:"action"`
	Number        int    `json:"number"`
	PullRequest   `json:"pull_request"`
	RequestedTeam `json:"requested_team"`
	Repository    `json:"repository"`
	Sender        `json:"sender"`
	Organization  `json:"organization"`
}

func now() string {
	return time.Now().Format(time.RFC3339)
}

func NewClient(ctx context.Context, appID int64, installationID int64, privateKeyFile string) (*github.Client, error) {

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, privateKeyFile)
	if err != nil {
		return nil, err
	}
	return github.NewClient(&http.Client{Transport: itr}), nil
}

func GetPREvents(ctx context.Context, client github.Client, logger *zap.Logger, req *http.Request) (eventIDs []string, msg message.Message, err error) {
	var body Webhook

	event := req.Header.Get("X-GitHub-Event")
	deliveryID := req.Header.Get("X-GitHub-Delivery")

	logFields := []zapcore.Field{
		zap.String("delivery_id", deliveryID),
		zap.String("event", event),
		zap.Int("number", body.Number),
		zap.String("repo", body.Repository.Name),
	}

	if event != "pull_request" {
		logger.Info("Skipped event",
			append(logFields,
				zap.String("event", "github.event_type.skipped"),
			)...)

		return eventIDs, msg, nil
	}

	err = json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		return eventIDs, msg, fmt.Errorf("failed to json decode request body: %w", err)
	}

	if body.Action != "review_requested" {
		logger.Info("Skipped event: unsupported event action",
			append(logFields,
				zap.String("event", "github.event_action.skipped"),
				zap.String("body_action", body.Action),
			)...)
		return eventIDs, msg, nil
	}

	if body.RequestedTeam.Name == "" && len(body.PullRequest.RequestedTeams) == 0 {
		logger.Info("Skipped event: team name missing",
			append(logFields,
				zap.String("event", "github.event_action.skipped"),
			)...)
		return eventIDs, msg, nil
	}

	logger.Info("Event action",
		append(logFields,
			zap.String("event", "github.event_action.found"),
			zap.String("body_action", body.Action),
		)...)

	eventIDs, htmlURL, err := getEventIds(ctx, client, body)
	if err != nil {
		return eventIDs, msg, fmt.Errorf("failed to get event IDs: %w", err)
	}

	msg = message.Message{
		URL:   htmlURL,
		Title: body.PullRequest.Title,
		Body:  body.PullRequest.Body,
	}

	return eventIDs, msg, nil
}

func getEventIds(ctx context.Context, client github.Client, body Webhook) ([]string, string, error) {
	var eventIDs []string
	var htmlURL string

	perPage := 100
	_, response, err := client.Issues.ListIssueTimeline(ctx, body.Organization.Login, body.Repository.Name, body.Number, &github.ListOptions{Page: 1, PerPage: perPage})
	if err != nil {
		return eventIDs, htmlURL, err
	}
	something := true

	currentPage := response.LastPage

	for something == true {
		var eventID string
		timeline, response, err := client.Issues.ListIssueTimeline(ctx, body.Organization.Login, body.Repository.Name, body.Number, &github.ListOptions{Page: currentPage, PerPage: perPage})
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

		}
	}

	pull, _, err := client.PullRequests.Get(ctx, body.Organization.Login, body.Repository.Name, body.Number)
	if err != nil {
		return eventIDs, htmlURL, err
	}
	htmlURL = pull.GetHTMLURL()

	return eventIDs, htmlURL, nil
}

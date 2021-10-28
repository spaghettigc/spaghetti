package formatmessage

import (
	"fmt"
	"os"
	"time"
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
type Installation struct { // app installation ID at org level
	ID     int    `json:"id"`
	NodeID string `json:"node_id"`
}
type Webhook struct {
	Action        string `json:"action"`
	Number        int    `json:"number"`
	PullRequest   `json:"pull_request"`
	RequestedTeam `json:"requested_team"`
	Repository    `json:"repository"`
	Sender        `json:"sender"`
	Installation  `json:"installation"`
}

func GetAssignee(body Webhook) (string, bool) {
	r := body.PullRequest.RequestedReviewers
	if len(r) == 0 { // TODO deal with potentially multiple teams
		return "", false
	}
	return r[0].Login, true
}

func FormatAssignee(githubUser string) string {
	cecile := os.Getenv("CECILE_SLACK_ID")
	jason := os.Getenv("JASON_SLACK_ID")

	userMap := map[string]string{
		"GitCecile": cecile,
		"thepesta":  jason,
	}

	slackUser, found := userMap[githubUser]

	assignee := githubUser
	if found {
		assignee = slackUser
	}

	return fmt.Sprintf(
		"<@%s> was assigned.",
		assignee,
	)
}

// - Cecile was assigneed. PR Title(reponame/#123)
// - BR is requested to review by someone
// - body
func FormatMessage(body Webhook) string {
	url := fmt.Sprintf("<%s|%s#%d>", body.PullRequest.HTMLURL, body.Repository.Name, body.Number)

	prTitle := fmt.Sprintf("PR title: %s (%s).", body.PullRequest.Title, url)

	teamAndSender := fmt.Sprintf(
		"%s team was requested to review by %s.",
		body.RequestedTeam.Name,
		body.Sender.Login,
	)

	prBody := fmt.Sprintf(
		"%s \n",
		Truncate(body.PullRequest.Body),
	)

	githubUser, assigned := GetAssignee(body)
	var assignee string
	if assigned {
		assignee = FormatAssignee(githubUser)
	}

	message := fmt.Sprintf(
		"%s %s\n"+
			"%s \n"+
			"%s",
		assignee,
		prTitle,
		teamAndSender,
		prBody,
	)

	return message
}

func Truncate(text string) string {
	size := 100
	if len(text) <= size {
		return text
	}
	return text[:size] + "..."
}

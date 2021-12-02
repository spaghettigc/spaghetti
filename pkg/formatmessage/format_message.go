package formatmessage

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
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

type Assigned struct {
	User string
	Team string
}

func GetAssignedReviewersAndTeam(resp *http.Response, eventID string) ([]Assigned, error) {
	assignees := make([]Assigned, 0)
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return assignees, err
	}
	doc.Find(fmt.Sprintf("#event-%s > div.TimelineItem-body", eventID)).Each(func(i int, s *goquery.Selection) {
		raw := s.Text()
		r, _ := regexp.Compile(`(?P<user>.+) \(assigned from (?P<team>.+)\)`)
		for _, txtArr := range r.FindAllStringSubmatch(raw, -1) {
			var user string
			var team string
			for i, t := range txtArr {
				if i == 0 { // entire match
					continue
				}
				if i == 1 { // user
					user = t
				}
				if i == 2 { // team
					team = t
				}
			}

			assignees = append(assignees, Assigned{User: user, Team: team})
		}
	})

	return assignees, nil
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
func FormatMessage(body Webhook, assignee Assigned) string {
	url := fmt.Sprintf("<%s|%s#%d>", body.PullRequest.HTMLURL, body.Repository.Name, body.Number)

	prTitle := fmt.Sprintf("PR title: %s (%s).", body.PullRequest.Title, url)

	teamAndSender := fmt.Sprintf(
		"%s team was requested to review by %s.",
		assignee.Team,
		body.Sender.Login,
	)

	prBody := fmt.Sprintf(
		"%s \n",
		Truncate(body.PullRequest.Body),
	)

	var assigneeMsg string
	assigneeMsg = FormatAssignee(assignee.User)

	message := fmt.Sprintf(
		"%s %s\n"+
			"%s \n"+
			"%s",
		assigneeMsg,
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

package formatmessage

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/go-rod/rod"
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

func GetAssignedReviewersAndTeam(eventID string, h string) ([]Assigned, error) {
	assignees := make([]Assigned, 0)

	// b, err := ioutil.ReadAll(resp.Body)
	// 5777536241
	// Request URL: https://github.com/spaghettigc/spaghetti/timeline_focused_item?after_cursor=Y3Vyc29yOnYyOpPPAAABfFpVwXABqjU0Mjc1ODgyNjc%3D&before_cursor=Y3Vyc29yOnYyOpPPAAABfXspOVgBqjU3MDU1NjQ0NDc%3D&id=PR_kwDOGB7j384scVH7&anchor=event-5777536241

	// err = ioutil.WriteFile(fmt.Sprintf("recording/%s.html", "boop"), b, 0644)

	// use headless browser
	page := rod.New().MustConnect().MustPage(h)
	time.Sleep(5 * time.Second) // TODO non timeout way to wait for client render
	// page.MustWaitLoad().MustScreenshot("a.png")

	selector := fmt.Sprintf("#event-%s > div.TimelineItem-body", eventID)
	el := page.MustElement(selector)
	text := el.MustText()

	fmt.Printf("\neventID: %s, text: %s\n", eventID, text)

	r, _ := regexp.Compile(`(?P<requester>.+) requested a review from (?P<reviewer>.+) \(assigned from (?P<team>.+)\)`)
	for _, txtArr := range r.FindAllStringSubmatch(text, -1) {
		var user string
		var team string
		for i, t := range txtArr {
			if i == 0 { // entire match
				continue
			}
			if i == 2 { // reviewer
				user = t
			}
			if i == 3 { // team
				team = t
			}
		}

		assignees = append(assignees, Assigned{User: user, Team: team})
	}

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

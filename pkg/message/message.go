package message

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/eko/gocache/marshaler"
	"github.com/go-rod/rod"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

type Message struct {
	URL   string
	Title string
	Body  string
}
type SlackOptions struct {
	Message   string
	ChannelID string
}

type PostMessageOptions struct {
	EventID     string
	ChannelID   string
	Logger      *zap.Logger
	Browser     *rod.Browser
	Marshal     *marshaler.Marshaler
	SlackClient *slack.Client
}

func PostMessage(options PostMessageOptions) error {
	cacheValue, err := options.Marshal.Get(options.EventID, new(Message))
	logger := options.Logger

	if err != nil {
		logger.Error("failed to get event ID from cache",
			zap.Error(err),
			zap.String("event_id", options.EventID),
		)

		return err
	}
	message := cacheValue.(*Message)
	url := fmt.Sprintf("%s#event-%s", message.URL, options.EventID)
	assignees, requester, err := GetAssignedReviewersAndTeam(options.Browser, options.EventID, url) // retunring zero assignees bug?
	if err != nil {
		logger.Error("failed to get assigned reviewers and team",
			zap.Error(err),
			zap.String("event_id", options.EventID),
			zap.String("url", message.URL),
			zap.String("full_url", url),
		)

		return err
	}

	logger.Info("identified assignees",
		zap.Int("assignee_count", len(assignees)),
		zap.String("event_id", options.EventID),
		zap.String("url", message.URL),
		zap.String("full_url", url),
	)

	for _, assignee := range assignees {

		slackMessage := FormatMessage(message.URL, message.Title, message.Body, assignee, requester)

		slackOptions := SlackOptions{
			Message:   slackMessage,
			ChannelID: options.ChannelID,
		}

		post(options.SlackClient, slackOptions)
	}

	return nil
}

func post(client *slack.Client, options SlackOptions) {
	msg := slack.MsgOptionText(options.Message, false)

	channelID, timestamp, err := client.PostMessage(options.ChannelID, msg)
	if err != nil {
		fmt.Printf("PostMessageErr: %s\n", err)
		// TODO return error
		return
	}
	fmt.Printf("channelID: %s, timestamp: %s", channelID, timestamp)
}

type Assigned struct {
	User string
	Team string
}

func FormatMessage(url string, title string, body string, assignee Assigned, requester string) string {
	prTitle := fmt.Sprintf("PR title: %s (%s).", title, url)

	teamAndSender := fmt.Sprintf(
		"%s team was requested to review by %s.",
		assignee.Team,
		requester,
	)

	prBody := fmt.Sprintf(
		"%s \n",
		Truncate(body),
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

func Truncate(text string) string {
	size := 100
	if len(text) <= size {
		return text
	}
	return text[:size] + "..."
}

// TODO rename this to include requester
func GetAssignedReviewersAndTeam(browser *rod.Browser, eventID string, h string) (assignees []Assigned, requester string, err error) {
	// assignees := make([]Assigned, 0)

	// b, err := ioutil.ReadAll(resp.Body)
	// 5777536241
	// Request URL: https://github.com/spaghettigc/spaghetti/timeline_focused_item?after_cursor=Y3Vyc29yOnYyOpPPAAABfFpVwXABqjU0Mjc1ODgyNjc%3D&before_cursor=Y3Vyc29yOnYyOpPPAAABfXspOVgBqjU3MDU1NjQ0NDc%3D&id=PR_kwDOGB7j384scVH7&anchor=event-5777536241

	// err = ioutil.WriteFile(fmt.Sprintf("recording/%s.html", "boop"), b, 0644)

	// use headless browser
	page := browser.MustPage(h)
	time.Sleep(1 * time.Second) // TODO non timeout way to wait for client render
	// page.MustWaitLoad().MustScreenshot("a.png")

	loaderElement := page.MustElement("#js-timeline-progressive-loader")
	timelineFocusedItem, err := loaderElement.Attribute("data-timeline-item-src")
	if err != nil {
		return assignees, requester, err
	}

	newUrl := fmt.Sprintf("https://github.com/%s&anchor=event-%s", *timelineFocusedItem, eventID)
	fmt.Printf("newUrl: %v\n", newUrl)

	newPage := browser.MustPage(newUrl)
	// htmlString, err := newPage.HTML()
	// fmt.Printf("htmlString: %s\n", htmlString)

	requestedAReviewFrom := newPage.MustElement("a[data-hovercard-type] > span") // This doesn't work as without authentication, the span doesn't appear
	requestedAReviewFromText := requestedAReviewFrom.MustText()
	fmt.Printf("requestedAReviewFromText: %s\n", requestedAReviewFromText)

	hoverType, err := requestedAReviewFrom.MustParent().Attribute("data-hovercard-type")
	fmt.Printf("hoverType: %s\n", *hoverType)

	selector := fmt.Sprintf("#event-%s > div.TimelineItem-body", eventID)
	el := newPage.MustElement(selector)
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

			if i == 1 {
				requester = strings.TrimSpace(t)
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

	return assignees, requester, nil
}

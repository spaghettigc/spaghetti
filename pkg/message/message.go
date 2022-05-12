package message

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/eko/gocache/marshaler"
	"github.com/getsentry/sentry-go"
	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	UserMapping UserMap
}

func PostMessage(options PostMessageOptions) error {
	cacheValue, err := options.Marshal.Get(options.EventID, new(Message))
	logger := options.Logger

	if err != nil {
		logger.Error("failed to get event ID from cache",
			zap.Error(err),
			zap.String("event_id", options.EventID),
		)

		return errors.Wrap(err, "failed to get event ID from cache")
	}
	message := cacheValue.(*Message)
	url := fmt.Sprintf("%s#event-%s", message.URL, options.EventID)
	timelineItemText, err := ScrapeTimelineItemText(options.Browser, options.EventID, url, options.Logger) // retunring zero assignees bug?
	assignees, requester, err := ParseTimelineItemText(timelineItemText)

	if err != nil {
		logger.Error("failed to get assigned reviewers and team",
			zap.Error(err),
			zap.String("event_id", options.EventID),
			zap.String("url", message.URL),
			zap.String("full_url", url),
		)

		return errors.Wrap(err, "failed to get assigned reviewers and team")
	}

	logger.Info("identified assignees",
		zap.Int("assignee_count", len(assignees)),
		zap.String("event_id", options.EventID),
		zap.String("url", message.URL),
		zap.String("full_url", url),
	)

	for _, assignee := range assignees {

		formatMessageOpt := FormatMessageOptions{
			URL:         message.URL,
			Title:       message.Title,
			Body:        message.Body,
			Assignee:    assignee,
			Requester:   requester,
			UserMapping: options.UserMapping,
		}
		slackMessage := FormatMessage(formatMessageOpt)

		slackOptions := SlackOptions{
			Message:   slackMessage,
			ChannelID: options.ChannelID,
		}

		err = post(options.SlackClient, slackOptions, logger)
	}

	return nil
}

func post(client *slack.Client, options SlackOptions, logger *zap.Logger) error {
	msg := slack.MsgOptionText(options.Message, false)

	logFields := []zapcore.Field{
		zap.String("channel_id", options.ChannelID),
		zap.String("msg_string", options.Message),
	}

	logger.Info("Attempting to post slack message",
		append(logFields,
			zap.String("event", "slack_message.post_started"),
		)...,
	)
	_, timestamp, err := client.PostMessage(options.ChannelID, msg)

	if err != nil {
		logger.Error("Failed to post slack message",
			append(logFields,
				zap.Error(err),
				zap.String("event", "slack_message.post_finished"),
				zap.String("outcome", "error"),
			)...,
		)
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Data: map[string]interface{}{
				"channel_id": options.ChannelID,
				"msg_string": options.Message,
			},
		})
		return errors.Wrap(err, "Failed to post slack message")
	}

	logger.Info("Successfully posted slack message",
		append(logFields,
			zap.String("event", "slack_message.post_finished"),
			zap.String("outcome", "success"),
			zap.String("slack_timestamp", timestamp),
		)...,
	)
	return nil
}

type Assigned struct {
	User string
	Team string
}

type UserMap map[string]string

type FormatMessageOptions struct {
	URL         string
	Title       string
	Body        string
	Assignee    Assigned
	Requester   string
	UserMapping UserMap
}

func FormatMessage(options FormatMessageOptions) string {
	prTitle := fmt.Sprintf("PR title: %s (%s).", options.Title, options.URL)

	teamAndSender := fmt.Sprintf(
		"%s team was requested to review by %s.",
		options.Assignee.Team,
		options.Requester,
	)

	prBody := fmt.Sprintf(
		"%s \n",
		Truncate(options.Body),
	)

	mappedAssignee, found := options.UserMapping[options.Assignee.User]

	if !found {
		mappedAssignee = options.Assignee.User
	}

	assigneeMsg := fmt.Sprintf(
		"<@%s> was assigned.",
		mappedAssignee,
	)

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

func ScrapeTimelineItemText(browser *rod.Browser, eventID string, h string, logger *zap.Logger) (string, error) {
	// use headless browser
	page := browser.MustPage(h)
	time.Sleep(1 * time.Second) // TODO non timeout way to wait for client render
	// page.MustWaitLoad().MustScreenshot("a.png")

	loaderElement := page.MustElement("#js-timeline-progressive-loader")
	timelineFocusedItem, err := loaderElement.Attribute("data-timeline-item-src")
	if err != nil {
		return "", errors.Wrap(err, "could not find timeline focused item")
	}

	newUrl := fmt.Sprintf("https://github.com/%s&anchor=event-%s", *timelineFocusedItem, eventID)

	logger.Info("Visiting github url",
		zap.String("event", "page_visit.started"),
		zap.String("url", newUrl),
	)

	newPage := browser.MustPage(newUrl)

	logger.Info("Finished visiting github url",
		zap.String("event", "page_visit.finished"),
		zap.String("url", newUrl),
	)

	logger.Info("Finding hovercard type",
		zap.String("event", "hovercard_type.search_started"),
	)

	selector := fmt.Sprintf("#event-%s > div.TimelineItem-body", eventID)

	logger.Info("Finding timeline item text",
		zap.String("event", "timeline_item.search_started"),
		zap.String("selector", selector),
		zap.String("event_id", eventID),
	)

	el := newPage.MustElement(selector)
	timelineItemText := el.MustText()

	logger.Info("Finding timeline item text",
		zap.String("event", "timeline_item.search_finished"),
		zap.String("selector", selector),
		zap.String("event_id", eventID),
		zap.String("text", timelineItemText),
	)

	return timelineItemText, nil
}

func ParseTimelineItemText(text string) (assignees []Assigned, requester string, err error) {

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

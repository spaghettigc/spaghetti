package message_test

import (
	"spaghetti/pkg/message"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Truncate_ShortMessage(t *testing.T) {
	g := NewGomegaWithT(t)
	text := "less than 100 chars"
	g.Expect(message.Truncate(text)).To(Equal(text))
}

func Test_Truncate_LongMessage(t *testing.T) {
	g := NewGomegaWithT(t)
	text := "long long long long long long long long long long long long long long long long long long long long long test: more than 100 chars"
	g.Expect(message.Truncate(text)).To(Equal(text[:100] + "..."))
}

func Test_FormatMessage_UserMappingSuccess(t *testing.T) {
	g := NewGomegaWithT(t)
	url := "https://github.com/dontgohere/pulls/9999/"
	title := "fix: wow we componentised pricing"
	body := "we did it"
	assignee := message.Assigned{
		User: "githubUser",
		Team: "Billing and revenue",
	}
	requester := "requester_user"
	userMapping := message.UserMap{
		"githubUser": "slackUser",
	}

	opt := message.FormatMessageOptions{
		URL:         url,
		Title:       title,
		Body:        body,
		Assignee:    assignee,
		Requester:   requester,
		UserMapping: userMapping,
	}
	message := message.FormatMessage(opt)
	g.Expect(message).To(Equal("<@slackUser> was assigned. PR title: fix: wow we componentised pricing (https://github.com/dontgohere/pulls/9999/).\nBilling and revenue team was requested to review by requester_user. \nwe did it \n"))
}

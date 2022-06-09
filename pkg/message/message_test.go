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

func Test_FormatMessage_UserMappingFailed(t *testing.T) {
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
		"notGithubUser": "slackUser",
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
	g.Expect(message).To(Equal("<@githubUser> was assigned. PR title: fix: wow we componentised pricing (https://github.com/dontgohere/pulls/9999/).\nBilling and revenue team was requested to review by requester_user. \nwe did it \n"))
}

func Test_ParseTimelineItemText_Success(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "requester_username requested a review from reviewer_username (assigned from team_name)"

	assignees, requester, err := message.ParseTimelineItemText(input)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(requester).Should(Equal("requester_username"))
	g.Expect(assignees).Should(Equal([]message.Assigned{
		{
			User: "reviewer_username",
			Team: "team_name",
		},
	}))
}

func Test_ParseTimelineItemText_Errors_WhenRequesterNotPresent(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "requested a review from reviewer_username (assigned from team_name)"

	_, _, err := message.ParseTimelineItemText(input)
	g.Expect(err).Should(HaveOccurred())
}

func Test_ParseTimelineItemText_Errors_WhenAssigneeNotPresent(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "requester_name requested a review from (assigned from team_name)"

	_, _, err := message.ParseTimelineItemText(input)
	g.Expect(err).Should(HaveOccurred())
}

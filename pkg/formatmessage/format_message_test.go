package formatmessage

import (
	"testing"
	"github.com/dnaeon/go-vcr/v2/recorder"
)
	

func TestGetAssigneeEmptyRequestedReviewers(t *testing.T) {
	body := Webhook{}
	have := GetAssignee(body)
	want := ""
	if have != want  {
		t.Fatalf("Have: %s, want: %s", have, want)
	}
}

func TestGetAssigneeHasRequestedReviewers(t *testing.T) {
	want := "someone"
	reviewers := []RequestedReviewer{
		{
			Login: want,
		},
	}

	body := Webhook{
		PullRequest: PullRequest{
			RequestedReviewers: reviewers,
		},
	}

	have := GetAssignee(body)
	if have != want {
		t.Fatalf("Have: %s, want: %s", have, want)
	}

func TestGetAssigneeHasMultipleRequestedReviewers(t *testing.T) {
	want := "someone"
	reviewers := []RequestedReviewer{
		{
			Login: "not me",
		},
		{
			Login: want,
		},
	}

	example_response.json

	body := Webhook{
		PullRequest: PullRequest{
			RequestedReviewers: reviewers,
		},
	}

	func getBodyFromReq() {
		fileWithReq := openRecording()

		body := fileWithReq.body

		return body
	}

	body := getBodyFromReq()

	have := GetAssignee(body)
	if have != want {
		t.Fatalf("Have: %s, want: %s", have, want)
	}
	recorder
}

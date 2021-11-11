package formatmessage

import "testing"

func TestGetAssigneeEmptyRequestedReviewers(t *testing.T) {
	body := Webhook{}
	have := GetAssignee(body)
	want := ""
	if have != want {
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
}

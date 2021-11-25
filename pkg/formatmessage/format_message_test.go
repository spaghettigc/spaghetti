package formatmessage

import (
	"encoding/json"
	"os"
	"reflect"
	"spaghetti/pkg/vcr"
	"testing"
)

func TestGetAssigneeEmptyRequestedReviewers(t *testing.T) {
	body := Webhook{}
	have := GetAssignees(body)
	want := []string{
		"",
	}
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("Have: %s, want: %s", have, want)
	}
}

func TestGetAssigneeHasRequestedReviewers(t *testing.T) {
	want := []string{"someone"}

	wd, _ := os.Getwd()
	filename := wd + "/recording/team-review-one-member.json"
	request, _ := vcr.ReadRequest(filename)

	var body Webhook

	err := json.Unmarshal([]byte(request.Body), &body)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	have := GetAssignees(body)
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("Have: %s, want: %s", have, want)
	}
}

func TestGetAssigneeHasMultipleRequestedReviewers(t *testing.T) {
	want := []string{"someone", "222"}

	wd, _ := os.Getwd()

	filename := wd + "/recording/review-multiple-members.json"
	request, _ := vcr.ReadRequest(filename)

	var body Webhook

	err := json.Unmarshal([]byte(request.Body), &body)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	have := GetAssignees(body)
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("Have: %s, want: %s", have, want)
	}

}

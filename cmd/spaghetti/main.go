package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// app_id, err := strconv.Atoi(os.Getenv("APP_ID"))
	app_id, err := strconv.ParseInt(os.Getenv("APP_ID"), 10, 64)
	if err != nil {
		log.Fatalf("app_id: %s", err)
	}
	installation_id, err := strconv.ParseInt(os.Getenv("INSTALLATION_ID"), 10, 64)
	if err != nil {
		log.Fatalf("app_id: %s", err)
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, app_id, installation_id, "spaghettiapp.2021-09-16.private-key.pem")
	if err != nil {
		log.Fatalf("key: %s", err)
	}
	client := github.NewClient(&http.Client{Transport: itr})

	pr, _, err := client.PullRequests.Get(ctx, "spaghettigc", "spaghetti", 3)
	fmt.Printf("%v\n", pr.RequestedReviewers)
	fmt.Printf("\n\n")
	bod := *pr.Body
	fmt.Printf("%s", bod)

	if err != nil {
		// this is currently failing with could not refresh installation id 138635's token:, we suspect it's something to do with installation ID, aka 2nd arg in ghinstallation.NewKeyFromFile
		log.Fatalf("client: %s", err)
	}

	// http.HandleFunc("/webhooks", func(w http.ResponseWriter, req *http.Request) {
	// 	var body Webhook
	// 	event := req.Header.Get("X-GitHub-Event")
	// 	deliveryID := req.Header.Get("X-GitHub-Delivery")
	// 	fmt.Printf("event: %s, deliveryID: %s\n", event, deliveryID)

	// 	err := json.NewDecoder(req.Body).Decode(&body)
	// 	if err != nil {
	// 		log.Fatalf("json decode: %s", err)
	// 	}
	// 	// fmt.Printf("%+v", body)

	// 	s, _ := json.MarshalIndent(body, "", "\t")
	// 	fmt.Print(string(s))
	// 	// body, err := io.ReadAll(req.Body)
	// })

	// fmt.Print("started serving")
	// err = http.ListenAndServe("localhost:3000", nil)
	// if err != nil {
	// 	log.Fatalf("http server: %s", err)
	// }
}

type Webhook struct {
	Action       string `json:"action"`
	Installation struct {
		ID      int `json:"id"`
		Account struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"account"`
		RepositorySelection string `json:"repository_selection"`
		AccessTokensURL     string `json:"access_tokens_url"`
		RepositoriesURL     string `json:"repositories_url"`
		HTMLURL             string `json:"html_url"`
		AppID               int    `json:"app_id"`
		AppSlug             string `json:"app_slug"`
		TargetID            int    `json:"target_id"`
		TargetType          string `json:"target_type"`
		Permissions         struct {
			Contents     string `json:"contents"`
			Metadata     string `json:"metadata"`
			PullRequests string `json:"pull_requests"`
		} `json:"permissions"`
		Events                 []interface{} `json:"events"`
		CreatedAt              time.Time     `json:"created_at"`
		UpdatedAt              time.Time     `json:"updated_at"`
		SingleFileName         interface{}   `json:"single_file_name"`
		HasMultipleSingleFiles bool          `json:"has_multiple_single_files"`
		SingleFilePaths        []interface{} `json:"single_file_paths"`
		SuspendedBy            interface{}   `json:"suspended_by"`
		SuspendedAt            interface{}   `json:"suspended_at"`
	} `json:"installation"`
	Sender struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"sender"`
}

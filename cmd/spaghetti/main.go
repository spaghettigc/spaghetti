package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 1, app_id, "spaghettiapp.2021-09-16.private-key.pem")
	if err != nil {
		log.Fatalf("key: %s", err)
	}
	client := github.NewClient(&http.Client{Transport: itr})
	repos, _, err := client.Repositories.List(ctx, "", nil)
	if err != nil {
		// this is currently failing with could not refresh installation id 138635's token:, we suspect it's something to do with installation ID, aka 2nd arg in ghinstallation.NewKeyFromFile
		log.Fatalf("client: %s", err)
	}
	fmt.Print(repos)
}

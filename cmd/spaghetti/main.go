package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	gh "spaghetti/pkg/github"
	"spaghetti/pkg/message"
	"strconv"
	"time"

	"os"

	"github.com/allegro/bigcache"
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/marshaler"
	"github.com/eko/gocache/store"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	slackAPI := slack.New(os.Getenv("SLACK_TOKEN"))
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	appID, err := strconv.ParseInt(os.Getenv("APP_ID"), 10, 64)
	if err != nil {
		panic(err)
	}
	installationID, err := strconv.ParseInt(os.Getenv("INSTALLATION_ID"), 10, 64)
	if err != nil {
		panic(err)
	}
	privateKeyFile := os.Getenv("PRIVATE_KEY_FILE")

	bigcacheClient, _ := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	bigcacheStore := store.NewBigcache(bigcacheClient, nil) // No options provided (as second argument
	cacheManager := cache.New(bigcacheStore)

	githubClient, err := gh.NewClient(ctx, appID, installationID, privateKeyFile)
	if err != nil {
		panic("GH client error")
	}

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, req *http.Request) {

		eventIDs, msg, err := gh.GetPREvents(ctx, *githubClient, req)
		if err != nil {
			panic("GH client error")
		}
		for _, eventID := range eventIDs {
			value, err := cacheManager.Get(eventID)

			if err != nil && err != bigcache.ErrEntryNotFound {
				panic(err)
			}

			if value != nil {
				fmt.Printf("skipped %s\n", eventID)
				continue
			}

			marshal := marshaler.New(cacheManager)

			err = marshal.Set(eventID, msg, nil)
			if err != nil {
				panic(err)
			}

			go message.PostMessage(marshal, slackAPI, channelID, eventID)
		}

	})

	fmt.Printf("started serving")
	err = http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		log.Fatalf("http server: %s", err)
	}
}

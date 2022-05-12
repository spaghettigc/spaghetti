package main

import (
	"context"
	"log"
	"net/http"
	"spaghetti/pkg/cache"
	gh "spaghetti/pkg/github"
	"spaghetti/pkg/message"
	"strconv"
	"time"

	"os"

	"github.com/allegro/bigcache"
	gocache "github.com/eko/gocache/cache"
	"github.com/eko/gocache/marshaler"
	"github.com/eko/gocache/store"
	"github.com/getsentry/sentry-go"
	"github.com/go-rod/rod"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// TODO error handling if we're stuck/can't load the page
func login(browser *rod.Browser, username string, password string) *rod.Browser {
	url := "https://github.com/login"
	page := browser.MustPage(url)

	page.MustElement("#login_field").MustInput(username)
	page.MustElement("#password").MustInput(password)
	time.Sleep(2 * time.Second)
	// page.MustScreenshot("before pressing.png")
	page.MustElement("#login > div.auth-form-body.mt-3 > form > div > input.btn.btn-primary.btn-block.js-sign-in-button").MustClick()
	time.Sleep(2 * time.Second)
	// page.MustScreenshot("login.png")
	return browser
}

func mainError(logger *zap.Logger) error {
	ctx := context.Background()

	// slack
	slackAPI := slack.New(os.Getenv("SLACK_TOKEN"))
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	appID, err := strconv.ParseInt(os.Getenv("APP_ID"), 10, 64)
	if err != nil {
		logger.Error("failed to parse app ID",
			zap.String("appID", os.Getenv("APP_ID")),
		)
		return errors.Wrap(err, "failed to parse app ID")
	}

	installationID, err := strconv.ParseInt(os.Getenv("INSTALLATION_ID"), 10, 64)
	if err != nil {
		logger.Error("failed to parse installation ID",
			zap.String("appID", os.Getenv("INSTALLATION_ID")),
		)
		return errors.Wrap(err, "failed to parse installation ID")
	}
	privateKeyFile := os.Getenv("PRIVATE_KEY_FILE")

	// cache
	bigcacheClient, err := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	if err != nil {
		logger.Error("failed to initialise bigcache client",
			zap.Error(err),
		)
		return errors.Wrap(err, "failed to initialise bigcache client")
	}
	bigcacheStore := store.NewBigcache(bigcacheClient, nil)
	cacheManager := gocache.New(bigcacheStore)
	marshal := marshaler.New(cacheManager)

	// github
	username := os.Getenv("GITHUB_NAME")
	password := os.Getenv("GITHUB_PWD")
	githubClient, err := gh.NewClient(ctx, appID, installationID, privateKeyFile)
	if err != nil {
		logger.Error("failed to initialise github client",
			zap.Error(err),
		)
		return errors.Wrap(err, "failed to initialise github client")
	}

	// headless browser
	browser := rod.New().MustConnect()
	// TODO what happens if we got logged out?
	browser = login(browser, username, password)

	// github <> slack mapping
	cecile := os.Getenv("CECILE_SLACK_ID")
	jason := os.Getenv("JASON_SLACK_ID")

	userMapping := message.UserMap{
		"GitCecile": cecile,
		"thepesta":  jason,
	}

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, req *http.Request) {

		eventIDs, msg, err := gh.GetPREvents(ctx, *githubClient, logger, req)
		if err != nil {
			logger.Error("failed to get the pr events",
				zap.Error(err),
			)
			err = errors.Wrap(err, "failed to get the pr events")
			sentry.CaptureException(err)
			return
		}

		unSeenEventIds, err := cache.ExcludeSeenEvents(logger, cacheManager, marshal, eventIDs, msg)

		if err != nil {
			logger.Error("failed to access event from cache",
				zap.Error(err),
			)
			err = errors.Wrap(err, "failed to access event from cache")
			sentry.CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		successfullyStoredIDs, failedToStoreIDs := cache.StoreInCache(logger, cacheManager, unSeenEventIds)

		for _, failedToStoreID := range failedToStoreIDs {
			err := failedToStoreID.Err

			sentry.AddBreadcrumb(&sentry.Breadcrumb{
				Data: map[string]interface{}{
					"event_id": failedToStoreID.EventId,
				},
			})

			sentry.CaptureException(err)
		}

		for _, eventID := range successfullyStoredIDs {
			opt := message.PostMessageOptions{
				EventID:     eventID,
				ChannelID:   channelID,
				Logger:      logger,
				Browser:     browser,
				Marshal:     marshal,
				SlackClient: slackAPI,
				UserMapping: userMapping,
			}

			// TODO what happens if postmessage fails?
			// we want to replace this
			go message.PostMessage(opt)
		}

	})

	logger.Info("started serving",
		zap.String("host", "localhost"),
		zap.Int("port", 3000),
	)
	// TODO configure host/port from env var/cli
	err = http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		logger.Error("failed to server",
			zap.Error(err),
		)
		return errors.Wrap(err, "failed to server")
	}
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	err = sentry.Init(sentry.ClientOptions{
		Environment: "",
		Release:     "",
		Dsn:         os.Getenv("SENTRY_DSN"),
		Debug:       true,
	})

	if err != nil {
		log.Fatalf("can't initialize sentry: %v", err)
	}

	defer sentry.Flush(time.Second * 2)
	defer sentry.Recover()

	err = mainError(logger)
	if err != nil {
		sentry.CaptureException(err)
	}
}

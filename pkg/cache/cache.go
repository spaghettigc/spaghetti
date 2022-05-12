package cache

import (
	"spaghetti/pkg/message"

	"github.com/allegro/bigcache"
	gocache "github.com/eko/gocache/cache"
	"github.com/eko/gocache/marshaler"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func ExcludeSeenEvents(logger *zap.Logger, cacheManager *gocache.Cache, marshal *marshaler.Marshaler, eventIDs []string, msg message.Message) ([]string, error) {
	unSeenEventIDs := []string{}
	for _, eventID := range eventIDs {
		value, err := cacheManager.Get(eventID)

		if err != nil && err != bigcache.ErrEntryNotFound {
			// accessing the cache isn't working
			logger.Error("failed to get event ID from cache",
				zap.Error(err),
				zap.String("event_id", eventID),
			)
			return unSeenEventIDs, errors.Wrap(err, "failed to access event from cache")
		}

		if value != nil {
			logger.Info("skipped the event ID as it's already in cache",
				zap.String("event_id", eventID),
			)
			continue
		}

		unSeenEventIDs = append(unSeenEventIDs, eventID)
	}
	return unSeenEventIDs, nil
}

type FailedIdsResult struct {
	EventId string
	Err     error
}

func StoreInCache(logger *zap.Logger, cacheManager *gocache.Cache, eventIDs []string) ([]string, []FailedIdsResult) {
	var successfullyStoredIDs []string // TODO replace with []EventIDs
	var failedToStoreIDs []FailedIdsResult

	for _, eventID := range eventIDs {
		err := cacheManager.Set(eventID, nil, nil)
		if err != nil {
			logger.Error("failed to marshal event ID",
				zap.Error(err),
				zap.String("event_id", eventID),
			)
			// sentry.AddBreadcrumb(&sentry.Breadcrumb{
			// 	Data: map[string]interface{}{
			// 		"event_id": eventID,
			// 	},
			// })
			err = errors.Wrap(err, "failed to marshal event ID")
			failedToStoreIDs = append(failedToStoreIDs, FailedIdsResult{EventId: eventID, Err: err})
			continue
			// sentry.CaptureException(err)
		}
		successfullyStoredIDs = append(successfullyStoredIDs, eventID)
	}

	return successfullyStoredIDs, failedToStoreIDs
}

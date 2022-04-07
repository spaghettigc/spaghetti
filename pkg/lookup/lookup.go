package lookup

import (
	"spaghetti/pkg/message"

	"github.com/allegro/bigcache"
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/marshaler"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func ExcludeSeenEvents(logger *zap.Logger, cacheManager *cache.Cache, marshal *marshaler.Marshaler, eventIDs []string, msg message.Message) []string {
	unSeenEventIDs := []string{}
	for _, eventID := range eventIDs {
		value, err := cacheManager.Get(eventID)

		if err != nil && err != bigcache.ErrEntryNotFound {
			// expected in two use cases
			// 1. haven't seen the event
			// 2. the cache has expired for the event
			logger.Error("failed to get event ID from cache",
				zap.Error(err),
				zap.String("event_id", eventID),
			)
		}

		if value != nil {
			logger.Info("skipped the event ID as it's already in cache",
				zap.String("event_id", eventID),
			)
			continue
		}

		err = marshal.Set(eventID, msg, nil)
		if err != nil {
			logger.Error("failed to marshal event ID",
				zap.Error(err),
				zap.String("event_id", eventID),
			)
			sentry.AddBreadcrumb(&sentry.Breadcrumb{
				Data: map[string]interface{}{
					"event_id": eventID,
				},
			})
			err = errors.Wrap(err, "failed to marshal event ID")
			sentry.CaptureException(err)
		}

		unSeenEventIDs = append(unSeenEventIDs, eventID)
	}
	return unSeenEventIDs
}

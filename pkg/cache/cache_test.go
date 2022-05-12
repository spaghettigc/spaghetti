package cache_test

import (
	"errors"
	"spaghetti/pkg/cache"
	"spaghetti/pkg/message"
	"testing"

	"github.com/allegro/bigcache"
	gocache "github.com/eko/gocache/cache"
	"github.com/eko/gocache/marshaler"
	mocksStore "github.com/eko/gocache/test/mocks/store"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func Test_UneventIDsCorrectlyReturnedWhenNotFound(t *testing.T) {
	g := NewGomegaWithT(t)

	logger := zap.NewExample()

	ctrl := gomock.NewController(t)
	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Get("event1").Return("event1", nil)
	store.EXPECT().Get("event2").Return(nil, bigcache.ErrEntryNotFound)

	c := gocache.New(store)
	marshal := marshaler.New(c)

	eventIDs := []string{"event1", "event2"}
	msg := message.Message{}
	unSeenEventIDs, err := cache.ExcludeSeenEvents(logger, c, marshal, eventIDs, msg)
	g.Expect(unSeenEventIDs).To(ConsistOf("event2"))
	g.Expect(err).ShouldNot(HaveOccurred())
}

func Test_CacheAccessFailed(t *testing.T) {
	g := NewGomegaWithT(t)

	logger := zap.NewExample()

	ctrl := gomock.NewController(t)
	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Get("event1").Return("event1", bigcache.ErrCannotRetrieveEntry)

	c := gocache.New(store)
	marshal := marshaler.New(c)
	eventIDs := []string{"event1"}
	msg := message.Message{}
	unSeenEventIDs, err := cache.ExcludeSeenEvents(logger, c, marshal, eventIDs, msg)
	g.Expect(unSeenEventIDs).To(BeEmpty())
	g.Expect(err).Should(HaveOccurred())
}

func Test_StoreInCacheFailed(t *testing.T) {
	g := NewGomegaWithT(t)

	logger := zap.NewExample()
	ctrl := gomock.NewController(t)
	store := mocksStore.NewMockStoreInterface(ctrl)
	errRandom := errors.New("bleh")
	store.EXPECT().Set("event1", nil, nil).Return(errRandom)
	eventIDs := []string{"event1"}
	c := gocache.New(store)

	successfullyStoredIDs, failedToStoreIDs := cache.StoreInCache(logger, c, eventIDs)
	g.Expect(successfullyStoredIDs).To(BeEmpty())
	g.Expect(failedToStoreIDs[0].Err).To(MatchError(errRandom))
	g.Expect(failedToStoreIDs[0].EventId).To(Equal("event1"))
}

func Test_StoreInCacheSuccess(t *testing.T) {
	g := NewGomegaWithT(t)

	logger := zap.NewExample()
	ctrl := gomock.NewController(t)
	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Set("event1", nil, nil).Return(nil)
	eventIDs := []string{"event1"}
	c := gocache.New(store)

	successfullyStoredIDs, failedToStoreIDs := cache.StoreInCache(logger, c, eventIDs)
	g.Expect(failedToStoreIDs).To(BeEmpty())
	g.Expect(successfullyStoredIDs).To(ConsistOf("event1"))
	g.Expect(len(successfullyStoredIDs)).To(Equal(1))
}

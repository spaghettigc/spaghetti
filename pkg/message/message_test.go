package message_test

import (
	"spaghetti/pkg/message"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_TruncateShortMessage(t *testing.T) {
	g := NewGomegaWithT(t)
	text := "less than 100 chars"
	g.Expect(message.Truncate(text)).To(Equal(text))
}

func Test_TruncateLongMessage(t *testing.T) {
	g := NewGomegaWithT(t)
	text := "long long long long long long long long long long long long long long long long long long long long long test: more than 100 chars"
	g.Expect(message.Truncate(text)).To(Equal(text[:100] + "..."))
}

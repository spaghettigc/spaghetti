package postmessage

import (
	"fmt"

	"github.com/slack-go/slack"
)

type PostMessageOptions struct {
	Message   string
	ChannelID string
}

func PostMessage(client *slack.Client, options PostMessageOptions) {

	msg := slack.MsgOptionText(options.Message, false)

	channelID, timestamp, err := client.PostMessage(options.ChannelID, msg)
	if err != nil {
		fmt.Printf("PostMessageErr: %s\n", err)
		return
	}
	fmt.Printf("channelID: %s, timestamp: %s", channelID, timestamp)

}

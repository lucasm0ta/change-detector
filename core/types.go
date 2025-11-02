package core

import (
	"net/url"
)

// Represents the request of a watcher of a certain website
type WatchRequest struct {
	// The channel information to return the feedback
	FeedbackChannelInfo *ChannelInfo `json:"id"`

	// URL to be checked if changed
	URL *url.URL `json:"url"`

	// interval in seconds, defaults to 3600s = 1h
	Interval int64
}

type WatchResponse struct {

	// URL to be checked if changed
	URL *url.URL `json:"url"`

	// External identification of the feedback channel. Like chatId, user email, etc.
	ChannelID string `json:"channel"`

	Diff string `json:"diff"`
}

type ChannelInfo struct {
	// Local identification of the client
	ID string `json:"id"`

	// Can be wither telegram, email, sms, or smoke_signal
	ChannelType string `json:"type"`

	// External identification of the feedback channel. Like chatId, user email, etc.
	ChannelID string `json:"channel"`
}

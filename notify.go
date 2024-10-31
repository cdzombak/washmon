package main

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/avast/retry-go"
	"github.com/cdzombak/gotfy"
)

func SendDoneNotification(ctx context.Context, cfg *Config, ackURL *url.URL, muteURL *url.URL) error {
	ntfyPublisher := gotfy.NewPublisher(gotfy.PublisherOpts{
		Server: cfg.NtfyServerURL(),
		Auth:   gotfy.AccessToken(cfg.NtfyToken),
		Headers: http.Header{
			"User-Agent": {productIdentifier()},
		},
	})

	return retry.Do(
		func() error {
			ctx, cancel := context.WithTimeout(ctx, cfg.NtfyTimeout())
			defer cancel()

			_, err := ntfyPublisher.Send(ctx, gotfy.Message{
				Topic:    cfg.NtfyTopic,
				Tags:     cfg.NtfyTags(),
				Priority: gotfy.Priority(cfg.NtfyPriority),
				Message:  "Washing machine is done.",
				Actions: []gotfy.ActionButton{
					&gotfy.HttpAction[string]{
						Label:  "✅ I emptied it",
						URL:    ackURL,
						Method: "POST",
						Clear:  true,
					},
					&gotfy.HttpAction[string]{
						Label:  "💤 Mute 3h",
						URL:    muteURL,
						Method: "POST",
						Clear:  true,
					},
				},
			})
			return err
		},
		retry.Attempts(2),
		retry.Delay(10*time.Second),
	)
}

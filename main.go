package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	_ "github.com/joho/godotenv/autoload"
)

type analytics struct {
	TotalViews int
}

func main() {
	var appHash string
	var appID int
	var a analytics
	appHash = os.Getenv("APP_HASH")
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatal(err)
	}
	channelUsername := "codative"
	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{Path: "user_session.json"},
	})
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		authenticator := termAuth{reader: bufio.NewReader(os.Stdin)}
		if err := client.Auth().IfNecessary(ctx, auth.NewFlow(authenticator, auth.SendCodeOptions{})); err != nil {
			return fmt.Errorf("auth: %w", err)
		}

		fmt.Println("--- Authenticated Successfully ---")

		api := client.API()
		resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
			Username: channelUsername,
		})
		if err != nil {
			return fmt.Errorf("resolve: %w", err)
		}

		if len(resolved.Chats) == 0 {
			return fmt.Errorf("channel not found")
		}

		channel, ok := resolved.Chats[0].(*tg.Channel)
		if !ok {
			return fmt.Errorf("peer is not a channel")
		}

		fmt.Printf("Channel: %s\n", channel.Title)
		startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
		minDateUnix := int(startDate.Unix())
		currentDate := int(time.Now().Unix())
		offsetID := 0
		offSet := currentDate
		limit := 100
		for offSet >= minDateUnix {
			peer := &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}
			res, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:       peer,
				OffsetDate: offSet,
				OffsetID:   offsetID,
				Limit:      limit,
			})
			if err != nil {
				return fmt.Errorf("history: %w. If you see BOT_METHOD_INVALID, delete old bot session and re-auth as a user (remove user_session.json)", err)
			}
			switch m := res.(type) {
			case *tg.MessagesChannelMessages:
				for _, msg := range m.Messages {
					if mm, ok := msg.(*tg.Message); ok {
						offSet = mm.Date
						a.TotalViews += mm.Views
					}
				}
			case *tg.MessagesMessages:
				for _, msg := range m.Messages {
					if mm, ok := msg.(*tg.Message); ok {
						offSet = mm.Date
						a.TotalViews += mm.Views
					}
				}
			case *tg.MessagesMessagesSlice:
				for _, msg := range m.Messages {
					if mm, ok := msg.(*tg.Message); ok {
						offSet = mm.Date
						a.TotalViews += mm.Views
					}
				}
			default:
				fmt.Printf("unexpected history type: %T\n", res)
			}
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("--- Total Views ---")
	fmt.Println(a.TotalViews)
}

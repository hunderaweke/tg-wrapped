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

func main() {
	var appHash string
	var appID int
	a := NewAnalytics()
	appHash = os.Getenv("APP_HASH")
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatal(err)
	}
	channelUsername := "Robi_makes_stuff"
	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{Path: "user_session.json"},
	})
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		authenticator := termAuth{reader: bufio.NewReader(os.Stdin)}
		if err := client.Auth().IfNecessary(ctx, auth.NewFlow(authenticator, auth.SendCodeOptions{})); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
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
		current := 1
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
			m, _ := res.(*tg.MessagesChannelMessages)
			for _, msg := range m.Messages {
				if mm, ok := msg.(*tg.Message); ok {
					if a.PopularPostID == 0 || a.PopularPostViewCount < mm.Views {
						a.PopularPostID = mm.ID
						a.PopularPostViewCount = mm.Views
					}
					if a.PopularPostByCommentID == 0 || a.PopularPostCommentCount < mm.Replies.Replies {
						a.PopularPostByCommentID = mm.ID
						a.PopularPostCommentCount = mm.Replies.Replies
					}
					offSet = mm.Date
					a.TotalViews += mm.Views
					a.TotalComments += mm.Replies.Replies
					t := getDateTime(mm.Date)
					a.MonthlyView[t.Month().String()] += mm.Views
					reactionCounter, totalReactions := (countNumOfReactions(mm.Reactions))
					a.ReactionCounter = mergeMaps(a.ReactionCounter, reactionCounter)
					a.TotalReactions += totalReactions
					a.PostCountPerMonth[t.Month().String()] += 1
					a.addDateCount(t)
					if fromID, ok := mm.FwdFrom.GetFromID(); ok {
						if ch, ok := fromID.(*tg.PeerChannel); ok {
							log.Printf("Forwarded: %d, Name: %s", ch.ChannelID, mm.FwdFrom.FromName)
							a.ForwardCount[int(ch.ChannelID)] += 1
						}
						a.TotalForwarded += 1
					}
				}
			}
			fmt.Printf("Current Loop: %d\n", current)
			current += 1
			time.Sleep(5 * time.Second)
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("--- Total Views ---")
	fmt.Printf("Total View: %d\n", a.TotalViews)
	fmt.Printf("Total Comments: %d\n", a.TotalComments)
	fmt.Printf("Total Reactions: %d\n", a.TotalReactions)
	// fmt.Println("Posts Per day: ", a.PostCountPerday)
	fmt.Printf("Max number of comments per post: %d\n", a.PopularPostCommentCount)
	fmt.Println("Total Forwarded Messages: ", a.TotalForwarded)
}

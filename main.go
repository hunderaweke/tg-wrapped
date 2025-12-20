package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	_ "github.com/joho/godotenv/autoload"
)

// termAuth implements the auth.UserAuthenticator interface to prompt for input via terminal.
type termAuth struct {
	reader *bufio.Reader
}

func (a termAuth) Phone(ctx context.Context) (string, error) {
	fmt.Print("Enter Phone (e.g. +1234567): ")
	s, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(s), nil
}

func (a termAuth) Password(ctx context.Context) (string, error) {
	fmt.Print("Enter 2FA Password (if any): ")
	s, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(s), nil
}

func (a termAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter Code: ")
	s, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(s), nil
}

// AcceptTermsOfService handles Telegram Terms of Service acceptance during authentication.
func (a termAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	fmt.Println("Telegram Terms of Service:")
	fmt.Println(tos.Text)
	fmt.Print("Do you accept the Terms of Service? (yes/no): ")
	s, _ := a.reader.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("terms of service not accepted")
	}
}

// SignUp collects user info for first-time registration.
func (a termAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	fmt.Print("Enter First Name: ")
	fn, _ := a.reader.ReadString('\n')
	fmt.Print("Enter Last Name: ")
	ln, _ := a.reader.ReadString('\n')
	return auth.UserInfo{
		FirstName: strings.TrimSpace(fn),
		LastName:  strings.TrimSpace(ln),
	}, nil
}

func main() {
	var appHash string
	var appID int
	// 1. Replace with your credentials
	appHash = os.Getenv("APP_HASH")
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatal(err)
	}
	channelUsername := "codative"
	client := telegram.NewClient(appID, appHash, telegram.Options{
		// Use a fresh session file to avoid reusing an old bot session
		SessionStorage: &telegram.FileSessionStorage{Path: "user_session.json"},
	})

	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// 2. Corrected Auth Flow using the terminal authenticator
		authenticator := termAuth{reader: bufio.NewReader(os.Stdin)}
		if err := client.Auth().IfNecessary(ctx, auth.NewFlow(authenticator, auth.SendCodeOptions{})); err != nil {
			return fmt.Errorf("auth: %w", err)
		}

		fmt.Println("--- Authenticated Successfully ---")

		api := client.API()

		// 3. Resolve Username
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

		// 5. Fetch recent messages (history)
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
			fmt.Println("--- Recent Messages ---")
			switch m := res.(type) {
			case *tg.MessagesChannelMessages:
				for _, msg := range m.Messages {
					if mm, ok := msg.(*tg.Message); ok {
						fmt.Printf("[%d] Views: %d Replies: %v ", mm.ID, mm.Views, mm.Replies.Replies)
						t := time.Unix(int64(mm.Date), 0).UTC().Format(time.RFC3339)
						fmt.Printf("Date: %s\n", t)
						offSet = mm.Date
					}
				}
			case *tg.MessagesMessages:
				for _, msg := range m.Messages {
					if mm, ok := msg.(*tg.Message); ok {
						fmt.Printf("[%d] Views: %d Replies: %v\n", mm.ID, mm.Views, mm.Replies.Replies)
						t := time.Unix(int64(mm.Date), 0).UTC().Format(time.RFC3339)
						fmt.Printf("Date: %s\n", t)
						offSet = mm.Date
					}
				}
			case *tg.MessagesMessagesSlice:
				for _, msg := range m.Messages {
					if mm, ok := msg.(*tg.Message); ok {
						fmt.Printf("[%d] Views: %d Replies: %v\n", mm.ID, mm.Views, mm.Replies.Replies)
						t := time.Unix(int64(mm.Date), 0).UTC().Format(time.RFC3339)
						fmt.Printf("Date: %s\n", t)
						offSet = mm.Date
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
}

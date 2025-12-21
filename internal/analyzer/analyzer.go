package analyzer

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	localAuth "github.com/hunderaweke/tg-unwrapped/internal/auth"
	_ "github.com/joho/godotenv/autoload"
)

type Analyzer struct {
	authenticator localAuth.TermAuth
	client        *telegram.Client
}

func NewAnalyzer() Analyzer {
	var appHash string
	var appID int
	appHash = os.Getenv("APP_HASH")
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatal(err)
	}
	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{Path: os.Getenv("APP_SESSION_STORAGE")},
	})
	authenticator := localAuth.NewTermAuth(bufio.NewReader(os.Stdin))
	return Analyzer{client: client, authenticator: authenticator}
}

func (ar *Analyzer) Run() error {
	stop, err := bg.Connect(ar.client)
	if err != nil {
		return err
	}
	defer func() { _ = stop() }()
	a, err := ar.ProcessAnalytics("codative")
	if err != nil {
		return err
	}
	fmt.Println("--- Total Views ---")
	fmt.Printf("Total View: %d\n", a.TotalViews)
	fmt.Printf("Total Comments: %d\n", a.TotalComments)
	fmt.Printf("Total Reactions: %d\n", a.TotalReactions)
	fmt.Printf("Max number of comments per post: %d\n", a.PopularPostCommentCount)
	fmt.Println("Total Forwarded Messages: ", a.TotalForwarded)
	if _, err := ar.client.Auth().Status(context.Background()); err != nil {
		return err
	}
	return nil
}

func (a *Analyzer) GetChannel(username string) (*tg.Channel, error) {
	var c *tg.Channel
	err := a.client.Auth().IfNecessary(context.Background(), auth.NewFlow(a.authenticator, auth.SendCodeOptions{}))
	if err != nil {
		return nil, fmt.Errorf("auth error: %w", err)
	}
	api := a.client.API()
	resolved, err := api.ContactsResolveUsername(context.Background(), &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, fmt.Errorf("resolving error: %w", err)
	}
	if len(resolved.Chats) == 0 {
		return nil, fmt.Errorf("channel not found")
	}
	c, ok := resolved.Chats[0].(*tg.Channel)
	if !ok {
		return nil, fmt.Errorf("chat is not channel")
	}
	return c, nil
}

func (ar *Analyzer) ProcessAnalytics(username string) (*Analytics, error) {
	channel, err := ar.GetChannel(username)
	if err != nil {
		return nil, err
	}
	a := NewAnalytics()
	api := ar.client.API()
	startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	minDateUnix := int(startDate.Unix())
	currentDate := int(time.Now().Unix())
	offsetID := 0
	offSet := currentDate
	limit := 100
	current := 1
	for offSet >= minDateUnix {
		peer := &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}
		res, err := api.MessagesGetHistory(context.Background(), &tg.MessagesGetHistoryRequest{
			Peer:       peer,
			OffsetDate: offSet,
			OffsetID:   offsetID,
			Limit:      limit,
		})
		if err != nil {
			return nil, fmt.Errorf("history: %w. If you see BOT_METHOD_INVALID, delete old bot session and re-auth as a user (remove user_session.json)", err)
		}
		m, _ := res.(*tg.MessagesChannelMessages)
		offSet = a.updateFromChannelMessages(m)
		fmt.Printf("Current Loop: %d\n", current)
		current += 1
		// *Important: need to deal with the rate limiter
		// time.Sleep(5 * time.Second)
	}
	return &a, err
}

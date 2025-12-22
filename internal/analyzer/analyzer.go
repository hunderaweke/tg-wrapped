package analyzer

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

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
	var a Analytics
	if err := ar.client.Run(context.Background(), func(ctx context.Context) error {
		channel, err := ar.GetChannel(username)
		a = NewAnalytics(channel.Title)
		if err != nil {
			return err
		}
		api := ar.client.API()
		peer := &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}
		type result struct {
			msg *tg.MessagesChannelMessages
			err error
		}
		currentDiff := 0
		now := time.Now()

		resultStream := make(chan result)
		go func() {
			done := make(chan int)
			defer close(done)
			defer close(resultStream)
			startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
			minDateUnix := int(startDate.Unix())
			finished := false
			curr := 1
			for !finished {
				var wg sync.WaitGroup
				for range 30 {
					wg.Add(1)
					go func(offSet int, done <-chan int) {
						var r result
						defer wg.Done()
						duration := time.Duration(offSet) * time.Hour
						t := int(now.Add(-duration).Unix())
						if t < minDateUnix {
							finished = true
						}
						for {
							res, err := api.MessagesGetHistory(context.Background(), &tg.MessagesGetHistoryRequest{
								Peer:       peer,
								OffsetDate: t,
							})
							if err != nil {
								r.err = fmt.Errorf("fetching history error: %w", err)
								time.Sleep(20 * time.Millisecond)
								continue
							}
							m, _ := res.(*tg.MessagesChannelMessages)
							r.msg = m
							break
						}
						select {
						case resultStream <- r:
						case <-done:
							return
						}

					}(currentDiff, done)
					currentDiff += 36
					time.Sleep(20 * time.Millisecond)
				}
				fmt.Println("Current : ", curr)
				wg.Wait()
				if finished {
					break
				}
				curr += 1
				time.Sleep(1 * time.Second)
			}
		}()
		for r := range resultStream {
			if r.err != nil {
				log.Println(r.err)
				continue
			}
			a.updateFromChannelMessages(r.msg)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	a.GetLongestStreak()
	return &a, nil
}

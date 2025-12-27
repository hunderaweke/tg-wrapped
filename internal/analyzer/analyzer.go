package analyzer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	localAuth "github.com/hunderaweke/tg-unwrapped/internal/auth"
	"github.com/hunderaweke/tg-unwrapped/internal/storage"
	_ "github.com/joho/godotenv/autoload"
)

type Analyzer struct {
	authenticator localAuth.TermAuth
	client        *telegram.Client
	minioClient   *storage.MinioClient
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
	minioBucket := os.Getenv("MINIO_BUCKET")
	minioClient, err := storage.NewMinioBucket(minioBucket)
	if err != nil {
		log.Fatal(err)
	}
	return Analyzer{client: client, authenticator: authenticator, minioClient: minioClient}
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
func (ar *Analyzer) DownloadProfile(c *tg.Channel) (string, error) {
	chatPhoto, ok := c.GetPhoto().(*tg.ChatPhoto)
	if !ok {
		return "", fmt.Errorf("error converting the photo")
	}
	location := &tg.InputPeerPhotoFileLocation{
		Peer: &tg.InputPeerChannel{
			AccessHash: c.AccessHash,
			ChannelID:  c.ID,
		},
		PhotoID: chatPhoto.PhotoID,
		Big:     true,
	}
	d := downloader.NewDownloader()
	var buf bytes.Buffer
	b := d.Download(ar.client.API(), location)
	_, err := b.Stream(context.Background(), &buf)
	if err != nil {
		return "", fmt.Errorf("error downloading the image: %w", err)
	}
	contentType := http.DetectContentType(buf.Bytes())
	fileExtensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(fileExtensions) == 0 {
		fileExtensions = []string{".jpg"}
	}
	fileName := fmt.Sprintf("%d%s", c.ID, fileExtensions[0])
	err = ar.minioClient.UploadProfile(fileName, buf, contentType)
	if err != nil {
		return "", fmt.Errorf("error uploading profile to minio: %w", err)
	}
	return fmt.Sprintf("/profiles/%s", fileName), nil
}
func (ar *Analyzer) ProcessAnalytics(username string) (*Analytics, error) {
	var a Analytics
	if err := ar.client.Run(context.Background(), func(ctx context.Context) error {
		channel, err := ar.GetChannel(username)
		if err != nil {
			return err
		}
		profileAddress, err := ar.DownloadProfile(channel)
		if err != nil {
			return err
		}
		a = NewAnalytics(channel.Title)
		a.ChannelProfile = profileAddress
		api := ar.client.API()
		startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
		minDateUnix := int(startDate.Unix())
		currentDate := int(time.Now().Unix())
		offsetID := 0
		offSet := currentDate
		limit := 100
		currentLoop := 1
		for offSet > minDateUnix {
			peer := &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}
			res, err := api.MessagesGetHistory(context.Background(), &tg.MessagesGetHistoryRequest{
				Peer:       peer,
				OffsetDate: offSet,
				OffsetID:   offsetID,
				Limit:      limit,
			})
			if err != nil {
				continue
			}
			m, ok := res.(*tg.MessagesChannelMessages)
			if !ok || m == nil {
				log.Printf("unexpected history response type: %T", res)
				offSet = minDateUnix
				continue
			}
			offSet = a.updateFromChannelMessages(m)
			currentLoop += 1
		}
		return nil
	}); err != nil {
		return nil, err
	}
	a.GetLongestStreak()
	return &a, nil
}

package analyzer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	localAuth "github.com/hunderaweke/tg-unwrapped/internal/auth"
	apperrors "github.com/hunderaweke/tg-unwrapped/internal/errors"
	"github.com/hunderaweke/tg-unwrapped/internal/logger"
	"github.com/hunderaweke/tg-unwrapped/internal/storage"
	_ "github.com/joho/godotenv/autoload"
)

const (
	defaultMessageLimit  = 100
	maxRetries           = 3
	retryDelay           = time.Second * 2
	defaultFileExtension = ".jpg"
)

type Analyzer struct {
	authenticator localAuth.TermAuth
	client        *telegram.Client
	minioClient   *storage.MinioClient
}

func NewAnalyzer(minioClient *storage.MinioClient) (*Analyzer, error) {
	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		return nil, apperrors.NewConfigError("APP_HASH", apperrors.ErrInvalidConfig)
	}

	appIDStr := os.Getenv("APP_ID")
	if appIDStr == "" {
		return nil, apperrors.NewConfigError("APP_ID", apperrors.ErrInvalidConfig)
	}

	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return nil, apperrors.NewConfigError("APP_ID", fmt.Errorf("invalid integer: %w", err))
	}

	sessionPath := os.Getenv("APP_SESSION_STORAGE")
	if sessionPath == "" {
		sessionPath = "session.json"
		logger.Warn("APP_SESSION_STORAGE not set, using default",
			"default", sessionPath)
	}

	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{Path: sessionPath},
	})

	authenticator := localAuth.NewTermAuth(bufio.NewReader(os.Stdin))

	logger.Info("Analyzer initialized successfully",
		"app_id", appID,
		"session_path", sessionPath)

	return &Analyzer{
		client:        client,
		authenticator: authenticator,
		minioClient:   minioClient,
	}, nil
}

func (a *Analyzer) GetChannel(ctx context.Context, username string) (*tg.Channel, error) {
	log := logger.With("operation", "GetChannel", "username", username)

	err := a.client.Auth().IfNecessary(ctx, auth.NewFlow(a.authenticator, auth.SendCodeOptions{}))
	if err != nil {
		log.Error("Authentication failed", "error", err)
		return nil, apperrors.NewAnalyzerError("auth", username, fmt.Errorf("%w: %v", apperrors.ErrAuthFailed, err))
	}

	api := a.client.API()
	resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		// Check if it's a "USERNAME_NOT_OCCUPIED" error from Telegram
		errStr := err.Error()
		if strings.Contains(errStr, "USERNAME_NOT_OCCUPIED") || strings.Contains(errStr, "username not occupied") {
			log.Warn("Username not found", "username", username)
			return nil, apperrors.NewAnalyzerError("resolve_username", username, apperrors.ErrChannelNotFound)
		}
		log.Error("Failed to resolve username", "error", err)
		return nil, apperrors.NewAnalyzerError("resolve_username", username, fmt.Errorf("%w: %v", apperrors.ErrTelegramAPI, err))
	}

	if len(resolved.Chats) == 0 {
		log.Warn("Channel not found")
		return nil, apperrors.NewAnalyzerError("resolve_username", username, apperrors.ErrChannelNotFound)
	}

	c, ok := resolved.Chats[0].(*tg.Channel)
	if !ok {
		log.Warn("Resolved chat is not a channel")
		return nil, apperrors.NewAnalyzerError("resolve_username", username, apperrors.ErrNotAChannel)
	}

	log.Info("Channel resolved successfully",
		"channel_id", c.ID,
		"channel_title", c.Title)

	return c, nil
}
func (ar *Analyzer) DownloadProfile(ctx context.Context, c *tg.Channel) (string, error) {
	if c == nil {
		return "", apperrors.NewAnalyzerError("download_profile", "", fmt.Errorf("channel is nil"))
	}

	log := logger.With("operation", "DownloadProfile", "channel_id", c.ID, "channel_title", c.Title)

	chatPhoto, ok := c.GetPhoto().(*tg.ChatPhoto)
	if !ok {
		log.Warn("Channel has no photo or invalid photo format")
		return "", apperrors.NewAnalyzerError("download_profile", c.Title, apperrors.ErrInvalidPhoto)
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

	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		b := d.Download(ar.client.API(), location)
		_, err = b.Stream(ctx, &buf)
		if err == nil {
			break
		}

		log.Warn("Download attempt failed",
			"attempt", attempt,
			"max_retries", maxRetries,
			"error", err)

		if attempt < maxRetries {
			time.Sleep(retryDelay * time.Duration(attempt))
			buf.Reset()
		}
	}

	if err != nil {
		log.Error("Failed to download profile after retries", "error", err)
		return "", apperrors.NewAnalyzerError("download_profile", c.Title, fmt.Errorf("%w: %v", apperrors.ErrDownloadFailed, err))
	}

	contentType := http.DetectContentType(buf.Bytes())
	fileExtensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(fileExtensions) == 0 {
		log.Debug("Could not determine file extension, using default",
			"content_type", contentType,
			"default_extension", defaultFileExtension)
		fileExtensions = []string{defaultFileExtension}
	}

	fileName := fmt.Sprintf("%d%s", c.ID, fileExtensions[0])
	err = ar.minioClient.UploadProfile(fileName, buf, contentType)
	if err != nil {
		log.Error("Failed to upload profile to storage", "error", err)
		return "", apperrors.NewAnalyzerError("upload_profile", c.Title, fmt.Errorf("%w: %v", apperrors.ErrUploadFailed, err))
	}

	profileURL := fmt.Sprintf("/profiles/%s", fileName)
	log.Info("Profile downloaded and uploaded successfully",
		"profile_url", profileURL,
		"content_type", contentType)

	return profileURL, nil
}
func (ar *Analyzer) fetchMessageDetails(ctx context.Context, api *tg.Client, channel *tg.Channel, messageID int) (*Message, error) {
	log := logger.With("operation", "fetchMessageDetails", "channel_id", channel.ID, "message_id", messageID)

	req := &tg.ChannelsGetMessagesRequest{
		Channel: channel.AsInput(),
		ID:      []tg.InputMessageClass{&tg.InputMessageID{ID: messageID}},
	}

	msg, err := api.ChannelsGetMessages(ctx, req)
	if err != nil {
		log.Error("Failed to fetch message", "error", err)
		return nil, apperrors.NewAnalyzerError("fetch_message", channel.Title, err)
	}

	channelMsg, ok := msg.(*tg.MessagesChannelMessages)
	if !ok || len(channelMsg.Messages) == 0 {
		log.Warn("No messages returned")
		return nil, apperrors.NewAnalyzerError("fetch_message", channel.Title, apperrors.ErrNoMessages)
	}

	tgMsg, ok := channelMsg.Messages[0].(*tg.Message)
	if !ok {
		log.Warn("Message is not a regular message")
		return nil, apperrors.NewAnalyzerError("fetch_message", channel.Title, apperrors.ErrNoMessages)
	}

	result := &Message{
		Text:  tgMsg.Message,
		Date:  getDateTime(tgMsg.Date),
		Views: tgMsg.Views,
	}

	if tgMsg.Replies.Replies != 0 {
		result.Comments = tgMsg.Replies.Replies
	}

	log.Debug("Message fetched successfully", "views", result.Views, "comments", result.Comments)
	return result, nil
}

func (ar *Analyzer) ProcessAnalytics(username string) (*Analytics, error) {
	log := logger.With("operation", "ProcessAnalytics", "username", username)
	log.Info("Starting analytics processing")

	startTime := time.Now()
	var a Analytics

	if err := ar.client.Run(context.Background(), func(ctx context.Context) error {
		channel, err := ar.GetChannel(ctx, username)
		if err != nil {
			return err
		}

		a = NewAnalytics(channel.Title)

		// Download channel profile (non-fatal if fails)
		profileAddress, err := ar.DownloadProfile(ctx, channel)
		if err != nil {
			log.Warn("Failed to download channel profile, continuing without it", "error", err)
		} else {
			a.ChannelProfile = profileAddress
		}

		api := ar.client.API()
		startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
		minDateUnix := int(startDate.Unix())
		currentDate := int(time.Now().Unix())
		offsetID := 0
		offSet := currentDate
		currentLoop := 1
		totalMessages := 0

		log.Info("Fetching channel messages",
			"channel", channel.Title,
			"start_date", startDate.Format("2006-01-02"))

		for offSet > minDateUnix {
			peer := &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}
			res, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:       peer,
				OffsetDate: offSet,
				OffsetID:   offsetID,
				Limit:      defaultMessageLimit,
			})
			if err != nil {
				log.Warn("Failed to fetch message batch, retrying",
					"loop", currentLoop,
					"error", err)
				time.Sleep(retryDelay)
				continue
			}

			m, ok := res.(*tg.MessagesChannelMessages)
			if !ok || m == nil {
				log.Debug("No more messages or invalid response", "loop", currentLoop)
				offSet = minDateUnix
				continue
			}

			messagesInBatch := len(m.Messages)
			totalMessages += messagesInBatch
			offSet = a.updateFromChannelMessages(m)

			log.Debug("Processed message batch",
				"loop", currentLoop,
				"messages_in_batch", messagesInBatch,
				"total_messages", totalMessages)

			currentLoop++
		}

		log.Info("Message fetching complete",
			"total_loops", currentLoop-1,
			"total_messages", totalMessages)

		// Fetch most viewed message details
		if a.Highlights.MostViewedID != 0 {
			mostViewed, err := ar.fetchMessageDetails(ctx, api, channel, a.Highlights.MostViewedID)
			if err != nil {
				log.Warn("Failed to fetch most viewed message details", "error", err)
			} else {
				a.Highlights.MostViewed = *mostViewed
			}
		}

		// Fetch most commented message details
		if a.Highlights.MostCommentedID != 0 {
			mostCommented, err := ar.fetchMessageDetails(ctx, api, channel, a.Highlights.MostCommentedID)
			if err != nil {
				log.Warn("Failed to fetch most commented message details", "error", err)
			} else {
				a.Highlights.MostCommented = *mostCommented
			}
		}

		// Download most forwarded channel profile (non-fatal if fails)
		if a.Highlights.MostForwardedChannel != nil {
			profileUrl, err := ar.DownloadProfile(ctx, a.Highlights.MostForwardedChannel)
			if err != nil {
				log.Warn("Failed to download forwarded channel profile",
					"forwarded_channel", a.Highlights.MostForwardedChannel.Title,
					"error", err)
			} else {
				a.Highlights.MostForwardedSource.Profile = profileUrl
			}
		}

		return nil
	}); err != nil {
		log.Error("Analytics processing failed", "error", err, "duration", time.Since(startTime))
		return nil, err
	}

	a.GetLongestStreak()

	log.Info("Analytics processing complete",
		"duration", time.Since(startTime),
		"total_posts", a.Totals.TotalPosts,
		"total_views", a.Totals.TotalViews)

	return &a, nil
}

package errors

import (
	"errors"
	"fmt"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
	ErrNotAChannel     = errors.New("chat is not a channel")
	ErrAuthFailed      = errors.New("authentication failed")
	ErrDownloadFailed  = errors.New("download failed")
	ErrUploadFailed    = errors.New("upload failed")
	ErrInvalidConfig   = errors.New("invalid configuration")
	ErrConnectionDead  = errors.New("connection dead")
	ErrNoMessages      = errors.New("no messages found")
	ErrInvalidPhoto    = errors.New("invalid photo format")
	ErrMinioConnection = errors.New("minio connection failed")
	ErrTelegramAPI     = errors.New("telegram API error")
	ErrRedisConnection = errors.New("redis connection failed")
)

type AnalyzerError struct {
	Op      string // Operation that failed
	Channel string // Channel being processed
	Err     error  // Underlying error
}

func (e *AnalyzerError) Error() string {
	if e.Channel != "" {
		return fmt.Sprintf("%s [channel=%s]: %v", e.Op, e.Channel, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *AnalyzerError) Unwrap() error {
	return e.Err
}

func NewAnalyzerError(op, channel string, err error) *AnalyzerError {
	return &AnalyzerError{
		Op:      op,
		Channel: channel,
		Err:     err,
	}
}

type ConfigError struct {
	Field string
	Err   error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error [field=%s]: %v", e.Field, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

func NewConfigError(field string, err error) *ConfigError {
	return &ConfigError{
		Field: field,
		Err:   err,
	}
}

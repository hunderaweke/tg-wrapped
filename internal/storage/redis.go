package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	apperrors "github.com/hunderaweke/tg-unwrapped/internal/errors"
	"github.com/hunderaweke/tg-unwrapped/internal/logger"
	"github.com/redis/go-redis/v9"
)

const (
	redisDefaultAddr     = "localhost:6379"
	redisDefaultPassword = ""
	redisDefaultDB       = 0
	redisTimeout         = 10 * time.Second
)

type RedisService struct {
	clnt *redis.Client
}

func NewRedis() (*RedisService, error) {
	log := logger.With("operation", "NewRedis")

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = redisDefaultAddr
		logger.Warn("REDIS_ADDR not set, using default", "default", addr)
	}

	password := os.Getenv("REDIS_PASSWORD")

	clnt := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       redisDefaultDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	stat := clnt.Ping(ctx)
	if stat.Err() != nil {
		log.Error("Failed to connect to Redis", "error", stat.Err())
		return nil, fmt.Errorf("%w: %v", apperrors.ErrRedisConnection, stat.Err())
	}

	logger.Info("Redis client initialized successfully", "addr", addr)
	return &RedisService{clnt: clnt}, nil
}

func (r *RedisService) Set(key string, v interface{}, expireTime time.Duration) error {
	log := logger.With("operation", "RedisSet", "key", key)

	data, err := json.Marshal(v)
	if err != nil {
		log.Error("Failed to marshal data", "error", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	cmd := r.clnt.Set(ctx, key, string(data), expireTime)
	_, err = cmd.Result()
	if err != nil {
		log.Error("Failed to set key", "error", err)
		return err
	}

	log.Debug("Key set successfully", "expiry", expireTime)
	return nil
}

func (r *RedisService) Get(key string, v interface{}) (bool, error) {
	log := logger.With("operation", "RedisGet", "key", key)

	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	cmd := r.clnt.Get(ctx, key)
	res, err := cmd.Result()
	if err == redis.Nil {
		log.Debug("Key not found")
		return false, nil
	}
	if err != nil {
		log.Error("Failed to get key", "error", err)
		return false, err
	}

	if err := json.Unmarshal([]byte(res), v); err != nil {
		log.Error("Failed to unmarshal data", "error", err)
		return false, err
	}

	log.Debug("Key retrieved successfully")
	return true, nil
}

func (r *RedisService) Delete(key string) error {
	log := logger.With("operation", "RedisDelete", "key", key)

	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	cmd := r.clnt.Del(ctx, key)
	if cmd.Err() != nil {
		log.Error("Failed to delete key", "error", cmd.Err())
		return cmd.Err()
	}

	log.Debug("Key deleted successfully")
	return nil
}

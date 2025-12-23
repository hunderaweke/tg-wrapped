package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	clnt *redis.Client
}

func NewRedis() (*RedisService, error) {
	clnt := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	stat := clnt.Ping(context.Background())
	if stat.Err() != nil {
		return nil, stat.Err()
	}
	return &RedisService{clnt: clnt}, nil
}

func (r *RedisService) Set(key string, v interface{}, expireTime time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	cmd := r.clnt.Set(context.Background(), key, string(data), expireTime)
	_, err = cmd.Result()
	if err != nil {
		return err
	}
	return nil
}
func (r *RedisService) Get(key string, v interface{}) (bool, error) {
	cmd := r.clnt.Get(context.Background(), key)
	res, err := cmd.Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(res), v); err != nil {
		return false, err
	}
	return true, nil
}
func (r *RedisService) Delete(key string) error {
	cmd := r.clnt.Del(context.Background(), key)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

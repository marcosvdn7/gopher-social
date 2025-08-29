package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"social/internal/store"
	"time"

	"github.com/go-redis/redis/v8"
)

const UserExpTime = time.Second * 30

type UserStore struct {
	client *redis.Client
}

func (s *UserStore) Get(ctx context.Context, userId int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userId)
	d, err := s.client.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var user store.User
	if d != "" {
		err := json.Unmarshal([]byte(d), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	if user.ID == 0 {
		return errors.New("missing info for redis key parsing")
	}
	log.Println("caching user in redis: ", user.ID)
	cacheKey := fmt.Sprintf("user-%v", user.ID)
	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.client.SetEX(ctx, cacheKey, json, UserExpTime).Err()
}

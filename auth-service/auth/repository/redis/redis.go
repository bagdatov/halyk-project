package rd

import (
	"fmt"
	"time"

	"auth-service/domain"

	"github.com/go-redis/redis"
)

type caсheStore struct {
	aTokenTTL time.Duration
	rTokenTTL time.Duration
	*redis.Client
}

func NewRedisClient(c *domain.Config) (domain.CaсheStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     c.CacheHost + c.CacheAddr,
		Password: c.CachePassword,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &caсheStore{
		aTokenTTL: c.AccessTokenTTL.Duration,
		rTokenTTL: c.RefreshTokenTTL.Duration,
		Client:    client,
	}, nil
}

func (c *caсheStore) FindToken(id int64, token string) bool {
	key := fmt.Sprintf("user:%d", id)

	value, err := c.Get(key).Result()
	if err != nil {
		return false
	}

	return token == value
}

func (c *caсheStore) InsertToken(id int64, token string) error {
	key := fmt.Sprintf("user:%d", id)

	return c.Set(key, token, c.rTokenTTL).Err()
}

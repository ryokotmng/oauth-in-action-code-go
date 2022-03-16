package pkg

import "github.com/go-redis/redis/v8"

type SessionRecord struct {
	AccessToken string `json:"access_token"`
}

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

package pkg

import "github.com/go-redis/redis/v8"

type TokenRecord struct {
	ClientID string `json:"client_id"`
}

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

package pkg

import "github.com/go-redis/redis/v8"

// TokenRecord keeps client id
type TokenRecord struct {
	ClientID string `json:"client_id"`
}

// NewRedisClient returns redis client with default options
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

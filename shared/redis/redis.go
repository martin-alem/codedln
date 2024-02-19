package redis

import (
	"github.com/redis/go-redis/v9"
	"os"
)

func ConnectToRedis() *redis.Client {
	// Define Redis options
	options := &redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PWD"),
		Username: os.Getenv("REDIS_USERNAME"),
	}
	// Create a new Redis client with the options
	rdb := redis.NewClient(options)

	return rdb
}

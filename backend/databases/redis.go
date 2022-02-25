package databases

import (
	"github.com/go-redis/redis/v7"
	_ "github.com/joho/godotenv/autoload"
	"os"
)

func NewRedisClient() *redis.Client {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URI"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}

	return redisClient
}

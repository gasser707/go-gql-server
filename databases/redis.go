package databases

import (
	"os"
	"github.com/go-redis/redis/v7"
	_ "github.com/joho/godotenv/autoload"
)


func NewRedisClient() *redis.Client  {

	 redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}

	return redisClient
}

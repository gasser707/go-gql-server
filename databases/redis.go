package databases

import (
	"os"
	"github.com/go-redis/redis/v7"
	_ "github.com/joho/godotenv/autoload"
)

var RedisClient *redis.Client

func init()  {

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := RedisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
}

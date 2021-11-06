package databases

import (
	// "os"
	"github.com/go-redis/redis/v7"
)

var RedisClient *redis.Client

func init()  {
	// redis_host := os.Getenv("REDIS_HOST")
	// redis_port := os.Getenv("REDIS_PORT")
	// redis_password := os.Getenv("REDIS_PASSWORD")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
}

package databases

import(
"os"
 "github.com/go-redis/redis/v7" 
)

var RC *redis.Client

func init() {
	//Initializing redis
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	RC = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := RC.Ping().Result()
	if err != nil {
		panic(err)
	}
}

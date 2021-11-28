package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gasser707/go-gql-server/databases"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/go-redis/redis/v7"
)

type RedisServiceInterface interface {
	CreateAuth(string, *TokenDetails) error
	FetchAuth(string) (string, error)
	DeleteRefresh(string) error
	DeleteTokens(*AccessDetails) error
}

type redisStoreService struct {
	client *redis.Client
}

var _ RedisServiceInterface = &redisStoreService{}

func NewRedisStore() *redisStoreService {
	return &redisStoreService{client: databases.NewRedisClient()}
}

//Save token metadata to Redis
func (tk *redisStoreService) CreateAuth(userId string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	atCreated, err := tk.client.Set(td.TokenUuid, userId, at.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	rtCreated, err := tk.client.Set(td.RefreshUuid, userId, rt.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	if atCreated == "0" || rtCreated == "0" {
		return customErr.Internal(context.Background(), "no record inserted")
	}
	return nil
}

//Check the metadata saved
func (tk *redisStoreService) FetchAuth(tokenUuid string) (string, error) {
	userid, err := tk.client.Get(tokenUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(context.Background(), err.Error())

	}
	return userid, nil
}

//Once a user row in the token table
func (tk *redisStoreService) DeleteTokens(authD *AccessDetails) error {
	//get the refresh uuid
	refreshUuid := fmt.Sprintf("%s++%s", authD.TokenUuid, authD.UserId)
	//delete access token
	deletedAt, err := tk.client.Del(authD.TokenUuid).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	//delete refresh token
	deletedRt, err := tk.client.Del(refreshUuid).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 {
		return customErr.Internal(context.Background(), "something went wrong")

	}
	return nil
}

func (tk *redisStoreService) DeleteRefresh(refreshUuid string) error {
	//delete refresh token
	deleted, err := tk.client.Del(refreshUuid).Result()
	if err != nil || deleted == 0 {
		return customErr.Internal(context.Background(), err.Error())
	}
	return nil
}

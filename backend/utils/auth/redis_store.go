package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gasser707/go-gql-server/databases"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/go-redis/redis/v7"
)

type RedisOperatorInterface interface {
	CreateAuth(string, *TokenDetails) error
	FetchAuth(tokenUuid string, csrfUuid string)  (string, error)
	FetchRefresh(refreshUuid string)  (string, error)
	DeleteRefresh(string) error
	DeleteTokens(*AccessDetails) error
}

type redisOperatorStore struct {
	client *redis.Client
}

var _ RedisOperatorInterface = &redisOperatorStore{}

func NewRedisStore() *redisOperatorStore {
	return &redisOperatorStore{client: databases.NewRedisClient()}
}

//Save token metadata to Redis
func (tk *redisOperatorStore) CreateAuth(userId string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	ct:= time.Unix(td.CsrfExpires, 0)
	now := time.Now()

	atCreated, err := tk.client.Set(td.TokenUuid, userId, at.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	rtCreated, err := tk.client.Set(td.RefreshUuid, userId, rt.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	csrfCreated, err := tk.client.Set(td.CsrfUuid, userId, ct.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	if atCreated == "0" || rtCreated == "0" || csrfCreated == "0" {
		return customErr.Internal(context.Background(), "no record inserted")
	}
	return nil
}

//Check the metadata saved
func (tk *redisOperatorStore) FetchAuth(tokenUuid string, csrfUuid string) (string, error) {
	userId, err := tk.client.Get(tokenUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(context.Background(), err.Error())

	}
	csrfUserId, err := tk.client.Get(csrfUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(context.Background(), err.Error())
	}
	if(userId!= csrfUserId){
		return "", customErr.NoAuth(context.Background(), err.Error())
	}
	return userId, nil
}

func (tk *redisOperatorStore) FetchRefresh(refreshUuid string)  (string, error){
	userId, err := tk.client.Get(refreshUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(context.Background(), err.Error())
	}
	return userId, nil
}


//Once a user row in the token table
func (tk *redisOperatorStore) DeleteTokens(authD *AccessDetails) error {
	//get the refresh uuid
	refreshUuid := fmt.Sprintf("%s++%s", authD.TokenUuid, authD.UserId)
	//delete access token
	deletedAt, err := tk.client.Del(authD.TokenUuid).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	deletedCsrf, err := tk.client.Del(authD.CsrfUuid).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	//delete refresh token
	deletedRt, err := tk.client.Del(refreshUuid).Result()
	if err != nil {
		return customErr.Internal(context.Background(), err.Error())
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 || deletedCsrf != 1 {
		return customErr.Internal(context.Background(), "something went wrong")

	}
	return nil
}

func (tk *redisOperatorStore) DeleteRefresh(refreshUuid string) error {
	//delete refresh token
	deleted, err := tk.client.Del(refreshUuid).Result()
	if err != nil || deleted == 0 {
		return customErr.Internal(context.Background(), err.Error())
	}
	return nil
}

package auth

import (
	"fmt"
	"github.com/gasser707/go-gql-server/databases"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/go-redis/redis/v7"
	"strings"
	"time"
)

type RedisOperatorInterface interface {
	CreateAuth(string, *TokenDetails) error
	FetchAuth(tokenUuid string, csrfUuid string) (string, error)
	FetchRefresh(refreshUuid string) (string, error)
	DeleteRefresh(string) error
	DeleteTokens(*AccessDetails) error
	DeleteAllUserTokens(userId string) error
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
	ct := time.Unix(td.CsrfExpires, 0)
	now := time.Now()

	atCreated, err := tk.client.Set(td.TokenUuid, userId, at.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	rtCreated, err := tk.client.Set(td.RefreshUuid, userId, rt.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	csrfCreated, err := tk.client.Set(td.CsrfUuid, userId, ct.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	if atCreated == "0" || rtCreated == "0" || csrfCreated == "0" {
		return customErr.Internal("no record inserted")
	}
	return nil
}

//Check the metadata saved
func (tk *redisOperatorStore) FetchAuth(tokenUuid string, csrfUuid string) (string, error) {
	userId, err := tk.client.Get(tokenUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(err.Error())

	}
	csrfUserId, err := tk.client.Get(csrfUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(err.Error())
	}
	if userId != csrfUserId {
		return "", customErr.NoAuth(err.Error())
	}
	return userId, nil
}

func (tk *redisOperatorStore) FetchRefresh(refreshUuid string) (string, error) {
	userId, err := tk.client.Get(refreshUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(err.Error())
	}
	return userId, nil
}

func (tk *redisOperatorStore) DeleteTokens(authD *AccessDetails) error {
	uuid := strings.Split(authD.TokenUuid, "@@")[0]
	//get the refresh uuid
	refreshUuid := fmt.Sprintf("%s++%s", uuid, authD.UserId)
	//delete access token
	deletedAt, err := tk.client.Del(authD.TokenUuid).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	deletedCsrf, err := tk.client.Del(authD.CsrfUuid).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	//delete refresh token
	deletedRt, err := tk.client.Del(refreshUuid).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 || deletedCsrf != 1 {
		return customErr.Internal("something went wrong")

	}
	return nil
}

func (tk *redisOperatorStore) DeleteAllUserTokens(userId string) error {
	var cursor uint64
	var keys []string
	iter := tk.client.Scan(cursor, fmt.Sprintf("*++%s", userId), 100).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}

	iter = tk.client.Scan(cursor, fmt.Sprintf("*@@%s", userId), 100).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}

	iter = tk.client.Scan(cursor, fmt.Sprintf("*$$%s", userId), 100).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}

	//delete refresh token
	deleted, err := tk.client.Del(keys...).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	//When the record is deleted, the return value is 1
	if deleted <= 0 {
		return customErr.Internal("something went wrong")

	}
	return nil

}

func (tk *redisOperatorStore) DeleteRefresh(refreshUuid string) error {
	//delete refresh token
	deleted, err := tk.client.Del(refreshUuid).Result()
	if err != nil || deleted == 0 {
		return customErr.Internal(err.Error())
	}
	return nil
}

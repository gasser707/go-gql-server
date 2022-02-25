package auth

import (
	"fmt"
	"github.com/gasser707/go-gql-server/databases"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/go-redis/redis/v7"
	"strings"
	"time"
)

type AuthStoreOperatorInterface interface {
	CreateAuthTokens(string, *TokenDetails) error
	FetchAuth(tokenUuid string, csrfUuid string) (string, error)
	FetchRefresh(refreshUuid string) (string, error)
	DeleteRefresh(string) error
	DeleteTokens(*AccessDetails) error
	DeleteAllUserTokens(userId string) error
}

type redisAuthStoreOperator struct {
	client *redis.Client
}

type authStoreOperator struct {
	authClient AuthStoreOperatorInterface
}

var _ AuthStoreOperatorInterface = &redisAuthStoreOperator{}

var _ AuthStoreOperatorInterface = &authStoreOperator{}

func NewRedisStore() *redisAuthStoreOperator {
	return &redisAuthStoreOperator{client: databases.NewRedisClient()}
}

func NewAuthStore(authClient AuthStoreOperatorInterface) *authStoreOperator {
	return &authStoreOperator{authClient}
}

func (as *authStoreOperator) CreateAuthTokens(userId string, td *TokenDetails) error {
	return as.authClient.CreateAuthTokens(userId, td)

}

func (as *authStoreOperator) FetchAuth(tokenUuid string, csrfUuid string) (string, error) {
	return as.authClient.FetchAuth(tokenUuid, csrfUuid)
}

func (as *authStoreOperator) FetchRefresh(refreshUuid string) (string, error) {
	return as.authClient.FetchRefresh(refreshUuid)
}

func (as *authStoreOperator) DeleteTokens(authD *AccessDetails) error {
	return as.authClient.DeleteTokens(authD)
}

func (as *authStoreOperator) DeleteAllUserTokens(userId string) error {
	return as.authClient.DeleteAllUserTokens(userId)
}

func (as *authStoreOperator) DeleteRefresh(refreshUuid string) error {
	return as.authClient.DeleteRefresh(refreshUuid)
}

//Save token metadata to Redis
func (rs *redisAuthStoreOperator) CreateAuthTokens(userId string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	ct := time.Unix(td.CsrfExpires, 0)
	now := time.Now()

	atCreated, err := rs.client.Set(td.TokenUuid, userId, at.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	rtCreated, err := rs.client.Set(td.RefreshUuid, userId, rt.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	csrfCreated, err := rs.client.Set(td.CsrfUuid, userId, ct.Sub(now)).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	if atCreated == "0" || rtCreated == "0" || csrfCreated == "0" {
		return customErr.Internal("no record inserted")
	}
	return nil
}

//Check the metadata saved
func (rs *redisAuthStoreOperator) FetchAuth(tokenUuid string, csrfUuid string) (string, error) {
	userId, err := rs.client.Get(tokenUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(err.Error())

	}
	csrfUserId, err := rs.client.Get(csrfUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(err.Error())
	}
	if userId != csrfUserId {
		return "", customErr.NoAuth(err.Error())
	}
	return userId, nil
}

func (rs *redisAuthStoreOperator) FetchRefresh(refreshUuid string) (string, error) {
	userId, err := rs.client.Get(refreshUuid).Result()
	if err != nil {
		return "", customErr.NoAuth(err.Error())
	}
	return userId, nil
}

func (rs *redisAuthStoreOperator) DeleteTokens(authD *AccessDetails) error {
	uuid := strings.Split(authD.TokenUuid, "@@")[0]
	//get the refresh uuid
	refreshUuid := fmt.Sprintf("%s++%s", uuid, authD.UserId)
	//delete access token
	deletedAt, err := rs.client.Del(authD.TokenUuid).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	deletedCsrf, err := rs.client.Del(authD.CsrfUuid).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	//delete refresh token
	deletedRt, err := rs.client.Del(refreshUuid).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 || deletedCsrf != 1 {
		return customErr.Internal("something went wrong")

	}
	return nil
}

func (rs *redisAuthStoreOperator) DeleteAllUserTokens(userId string) error {
	var cursor uint64
	var keys []string
	iter := rs.client.Scan(cursor, fmt.Sprintf("*++%s", userId), 100).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}

	iter = rs.client.Scan(cursor, fmt.Sprintf("*@@%s", userId), 100).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}

	iter = rs.client.Scan(cursor, fmt.Sprintf("*$$%s", userId), 100).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}

	//delete refresh token
	deleted, err := rs.client.Del(keys...).Result()
	if err != nil {
		return customErr.Internal(err.Error())
	}
	//When the record is deleted, the return value is 1
	if deleted <= 0 {
		return customErr.Internal("something went wrong")

	}
	return nil

}

func (rs *redisAuthStoreOperator) DeleteRefresh(refreshUuid string) error {
	//delete refresh token
	deleted, err := rs.client.Del(refreshUuid).Result()
	if err != nil || deleted == 0 {
		return customErr.Internal(err.Error())
	}
	return nil
}

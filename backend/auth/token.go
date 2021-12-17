package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gorilla/securecookie"
	"github.com/twinj/uuid"
)

type AccessDetails struct {
	TokenUuid string
	UserId    string
	UserRole  model.Role
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	TokenUuid    string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type TokenOperatorInterface interface {
	CreateToken(userId string, userRole model.Role) (*TokenDetails, error)
	ExtractTokenMetadata(c context.Context) (*AccessDetails, error)
}

//Token implements the TokenInterface
var _ TokenOperatorInterface = &tokenOperator{}

type tokenOperator struct {
	sc *securecookie.SecureCookie
}

func NewTokenOperator(sc *securecookie.SecureCookie) *tokenOperator {
	return &tokenOperator{
		sc: sc,
	}
}

func (t *tokenOperator) CreateToken(userId string, userRole model.Role) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 30).Unix() //expires after 30 min
	td.TokenUuid = uuid.NewV4().String()
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.TokenUuid + "++" + userId

	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.TokenUuid
	atClaims["user_id"] = userId
	atClaims["exp"] = td.AtExpires
	atClaims["user_role"] = userRole
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, customErr.Internal(context.Background(), err.Error())
	}

	//Creating Refresh Token
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.TokenUuid + "++" + userId

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userId
	rtClaims["exp"] = td.RtExpires
	rtClaims["user_role"] = userRole

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)

	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, customErr.Internal(context.Background(), err.Error())
	}
	return td, nil
}

func (t *tokenOperator) verifyToken(ctx context.Context) (*jwt.Token, error) {
	tokenMap, err := t.getTokensFromCookie(ctx)
	if err != nil {
		return nil, err
	}
	tokenString := tokenMap["access_token"]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, customErr.NoAuth(ctx, err.Error())

	}
	return token, nil
}

func extract(token *jwt.Token) (*AccessDetails, error) {

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		userId, userOk := claims["user_id"].(string)
		role, roleOk := claims["user_role"].(string)
		if !ok || !userOk || !roleOk {
			return nil, customErr.NoAuth(context.Background(), "unauthorized")

		} else {
			return &AccessDetails{
				TokenUuid: accessUuid,
				UserId:    userId,
				UserRole:  model.Role(role),
			}, nil
		}
	}
	return nil, customErr.NoAuth(context.Background(), "something went wrong")

}

func (t *tokenOperator) ExtractTokenMetadata(ctx context.Context) (*AccessDetails, error) {
	token, err := t.verifyToken(ctx)
	if err != nil {
		return nil, customErr.NoAuth(ctx, err.Error())
	}
	acc, err := extract(token)
	if err != nil {
		return nil, customErr.NoAuth(ctx, err.Error())
	}
	return acc, nil
}

func (t *tokenOperator) getTokensFromCookie(ctx context.Context) (map[string]string, error) {
	ca, err := GetCookieAccess(ctx)
	if err != nil {
		return nil, err
	}
	ec := ca.encodedCookie
	value := make(map[string]string)
	if err = t.sc.Decode("cookie-name", ec, &value); err != nil {
		return nil, customErr.NoAuth(ctx, err.Error())
	}

	return value, nil
}

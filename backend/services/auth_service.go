package services

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gasser707/go-gql-server/auth"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AuthServiceInterface interface {
	Login(ctx context.Context, input model.LoginInput) (bool, error)
	validateCredentials(c context.Context) (intUserID, model.Role, error)
	Logout(c context.Context) (bool, error)
	Refresh(c *gin.Context)
}

//UsersService implements the usersServiceInterface
var _ AuthServiceInterface = &authService{}

type authService struct {
	rd auth.RedisServiceInterface
	tk auth.TokenServiceInterface
	DB *sql.DB
	sc *securecookie.SecureCookie
}

func NewAuthService(db *sql.DB) *authService {
	sc := helpers.NewSecureCookie()
	tk := auth.NewTokenService(sc)
	rd := auth.NewRedisStore()
	return &authService{rd, tk, db, sc}
}

type UserID string
type intUserID int64

func (s *authService) Login(ctx context.Context, input model.LoginInput) (bool, error) {

	user, err := dbModels.Users(Where("email = ?", input.Email)).One(ctx, s.DB)
	if err != nil {
		return false, customErr.NoAuth(ctx, err.Error())
	}

	ok := helpers.CheckPasswordHash(input.Password, user.Password)
	if !ok || err != nil {
		return false, customErr.BadRequest(ctx, err.Error())
	}

	id := fmt.Sprintf("%v", user.ID)
	role := fmt.Sprintf("%v", user.Role)

	ts, err := s.tk.CreateToken(id, model.Role(role))
	if err != nil {
		return false, err
	}
	saveErr := s.rd.CreateAuth(id, ts)
	if saveErr != nil {
		return false, err
	}

	ca, err := auth.GetCookieAccess(ctx)
	if err != nil {
		return false, err
	}
	ca.SetToken(ts.AccessToken, ts.RefreshToken, s.sc)
	return true, nil
}

func (s *authService) validateCredentials(ctx context.Context) (intUserID, model.Role, error) {
	metadata, err := s.tk.ExtractTokenMetadata(ctx)
	if err != nil {
		return -1, "", err
	}
	userId, err := s.rd.FetchAuth(metadata.TokenUuid)
	if err != nil {
		return -1, "", err
	}

	id, err := strconv.Atoi(userId)
	if err != nil {
		return -1, "", customErr.Internal(ctx, err.Error())
	}

	return intUserID(id), metadata.UserRole, nil
}

func (s *authService) Logout(ctx context.Context) (bool, error) {
	//If metadata is passed and the tokens valid, delete them from the redis store
	metadata, err := s.tk.ExtractTokenMetadata(ctx)
	if err != nil {
		return false, err
	}
	err = s.rd.DeleteTokens(metadata)
	if err != nil {
		return false, err
	}
	return true, nil

}

func (s *authService) Refresh(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	refreshToken := mapToken["refresh_token"]

	//verify the token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		c.JSON(http.StatusUnauthorized, "Refresh token expired")
		return
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		c.JSON(http.StatusUnauthorized, err)
		return
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		userId, userOk := claims["user_id"].(string)
		userRole, roleOk := claims["user_role"].(string)
		if !roleOk || !userOk {
			c.JSON(http.StatusUnprocessableEntity, "unauthorized")
			return
		}
		//Delete the previous Refresh Token
		delErr := s.rd.DeleteRefresh(refreshUuid)
		if delErr != nil { //if any goes wrong
			c.JSON(http.StatusUnauthorized, "unauthorized")
			return
		}
		//Create new pairs of refresh and access tokens
		ts, createErr := s.tk.CreateToken(userId, model.Role(userRole))
		if createErr != nil {
			c.JSON(http.StatusForbidden, createErr.Error())
			return
		}
		//save the tokens metadata to redis
		saveErr := s.rd.CreateAuth(userId, ts)
		if saveErr != nil {
			c.JSON(http.StatusForbidden, saveErr.Error())
			return
		}
		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}
		c.JSON(http.StatusCreated, tokens)
	} else {
		c.JSON(http.StatusUnauthorized, "refresh expired")
	}
}

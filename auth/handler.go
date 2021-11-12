package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	db "github.com/gasser707/go-gql-server/databases"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/gin-gonic/gin"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// ProfileHandler struct
type profileHandler struct {
	rd AuthInterface
	tk TokenInterface
}

func NewProfile(rd AuthInterface, tk TokenInterface) *profileHandler {
	return &profileHandler{rd, tk}
}

var AuthService *profileHandler

type UserID string
type intUserID int64

func init() {
	var rd = NewRedisStore(db.RedisClient)
	var tk = NewToken()
	AuthService = NewProfile(rd, tk)
}

type Todo struct {
	UserID string `json:"user_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func (h *profileHandler) Login(ctx context.Context, input model.LoginInput) (bool, error) {

	user, err := dbModels.Users(Where("email = ?",input.Email)).One(ctx, db.MysqlDB)
	if(err!=nil){
		return false, err
	}
	ok:= helpers.CheckPasswordHash(input.Password, user.Password)

	if(!ok){
		return false, fmt.Errorf("wrong email password combination")
	}
	id := fmt.Sprintf("%v",user.ID)
	role := fmt.Sprintf("%v",user.Role)

	ts, err := AuthService.tk.CreateToken(id, model.Role(role))
	if err != nil {
		return false, err
	}
	saveErr := AuthService.rd.CreateAuth(id, ts)
	if saveErr != nil {
		return false, err
	}

	ca, err := GetCookieAccess(ctx)
	if err != nil {
		return false, err
	}
	ca.SetToken(ts.AccessToken, ts.RefreshToken)
	return true, nil
}


func (h *profileHandler) validateCredentials(c context.Context) (UserID, model.Role, error){
	metadata, err := h.tk.ExtractTokenMetadata(c)
	if(err !=nil){
		return "", "",err
	}
	userId, err := h.rd.FetchAuth(metadata.TokenUuid)
	if err != nil {
		return "", "",err
	}

	return UserID(userId), metadata.UserRole ,nil
}

func (h *profileHandler) Logout(c context.Context) (bool, error) {
	//If metadata is passed and the tokens valid, delete them from the redis store
	metadata, _ := h.tk.ExtractTokenMetadata(c)
	deleteErr := h.rd.DeleteTokens(metadata)
	if deleteErr != nil {
		return false, deleteErr
	}
	return true, nil

}


func (h *profileHandler) GetCredentials(c context.Context) (intUserID, error){
	metadata, _ := h.tk.ExtractTokenMetadata(c)
	userId, _ := h.rd.FetchAuth(metadata.TokenUuid)
	id, err := strconv.Atoi(string(userId))
	if err != nil {
		return -1, err
	}
	return intUserID(id), nil
}


// func (h *profileHandler) CreateTodo(c *gin.Context) {
// 	var td Todo
// 	if err := c.ShouldBindJSON(&td); err != nil {
// 		c.JSON(http.StatusUnprocessableEntity, "invalid json")
// 		return
// 	}
// 	metadata, err := h.tk.ExtractTokenMetadata(c)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, "unauthorized")
// 		return
// 	}
// 	userId, err := h.rd.FetchAuth(metadata.TokenUuid)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, "unauthorized")
// 		return
// 	}
// 	td.UserID = userId

// 	//you can proceed to save the  to a database

// 	c.JSON(http.StatusCreated, td)
// }

func (h *profileHandler) Refresh(c *gin.Context) {
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
		delErr := h.rd.DeleteRefresh(refreshUuid)
		if delErr != nil { //if any goes wrong
			c.JSON(http.StatusUnauthorized, "unauthorized")
			return
		}
		//Create new pairs of refresh and access tokens
		ts, createErr := h.tk.CreateToken(userId, model.Role(userRole))
		if createErr != nil {
			c.JSON(http.StatusForbidden, createErr.Error())
			return
		}
		//save the tokens metadata to redis
		saveErr := h.rd.CreateAuth(userId, ts)
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

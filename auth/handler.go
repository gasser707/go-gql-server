package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gasser707/go-gql-server/databases"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gin-gonic/gin"
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

func init() {
	var rd = NewRedisStore(databases.RedisClient)
	var tk = NewToken()
	AuthService = NewProfile(rd, tk)
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     model.Role `json:"role"`
	Email string `json:"Email"`
}

//In memory user
var user = User{
	ID:       "1",
	Username: "username",
	Password: "password",
	Email: "g@g.com",
	Role:     model.RoleUser,
}

type Todo struct {
	UserID string `json:"user_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func (h *profileHandler) Login(c context.Context) (bool, error) {
	// var u User
	// if err := c.ShouldBindJSON(&u); err != nil {
	// 	c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
	// 	return
	// }
	//compare the user from the request, with the one we defined:
	// if user.Username != u.Username || user.Password != u.Password {
	// 	c.JSON(http.StatusUnauthorized, "Please provide valid login details")
	// 	return
	// }
	ts, err := AuthService.tk.CreateToken(user.ID, user.Role)
	if err != nil {
		return false, err
	}
	saveErr := AuthService.rd.CreateAuth(user.ID, ts)
	if saveErr != nil {
		return false, err
	}

	ca, err := GetCookieAccess(c)
	if err != nil {
		return false, err
	}
	ca.SetToken(ts.AccessToken, ts.RefreshToken)
	return true, nil
}


func (h *profileHandler) validateCredentials(c context.Context) (string, model.Role, error){
	metadata, err := h.tk.ExtractTokenMetadata(c)
	if(err !=nil){
		return "", "",err
	}
	userId, err := h.rd.FetchAuth(metadata.TokenUuid)
	if err != nil {
		return "", "",err
	}

	return userId, metadata.UserRole ,nil
}

func (h *profileHandler) GetCredentials(c context.Context) (*AccessDetails, error){
		return h.tk.ExtractTokenMetadata(c)
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

func (h *profileHandler) CreateTodo(c *gin.Context) {
	var td Todo
	if err := c.ShouldBindJSON(&td); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid json")
		return
	}
	metadata, err := h.tk.ExtractTokenMetadata(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	userId, err := h.rd.FetchAuth(metadata.TokenUuid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	td.UserID = userId

	//you can proceed to save the  to a database

	c.JSON(http.StatusCreated, td)
}

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

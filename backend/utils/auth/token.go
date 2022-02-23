package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graphql/model"
	"github.com/gasser707/go-gql-server/middleware"
	"github.com/gorilla/securecookie"
	"github.com/twinj/uuid"
)

var (
	accessSecret        = os.Getenv("ACCESS_SECRET")
	refreshSecret       = os.Getenv("REFRESH_SECRET")
	csrfSecret          = os.Getenv("CSRF_SECRET")
	validationSecret    = os.Getenv("VALIDATION_SECRET")
	passwordResetSecret = os.Getenv("PASSWORD_RESET_SECRET")
)

type AccessDetails struct {
	TokenUuid string
	CsrfUuid  string
	UserId    string
	UserRole  model.Role
}

type RefreshDetails struct {
	RefreshUuid string
	UserId      string
	Role        model.Role
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	CsrfToken    string
	BaseUuid     string
	TokenUuid    string
	RefreshUuid  string
	CsrfUuid     string
	AtExpires    int64
	RtExpires    int64
	CsrfExpires  int64
}

type StatelessToken string

const (
	ResetToken        StatelessToken = "RESET_PASSWORD"
	ValidateUserToken StatelessToken = "VALIDATE_USER"
)

type TokenOperatorInterface interface {
	CreateTokens(userId string, userRole model.Role) (*TokenDetails, error)
	ExtractTokenMetadata(c context.Context) (*AccessDetails, error)
	ExtractRefreshMetadata(ctx context.Context) (*RefreshDetails, error)
	ExtractStatelessTokenMetadata(ctx context.Context, tokenString string, kind StatelessToken) (string, error)
	CreateStatelessToken(userId string, kind StatelessToken) (string, error)
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

func (t *tokenOperator) createCsrfToken(userId string, td *TokenDetails) (*TokenDetails, error) {
	td.CsrfExpires = time.Now().Add(time.Minute * 30).Unix() //expires after 30 min
	td.CsrfUuid = td.BaseUuid + "$$" + userId

	csrfClaims := jwt.MapClaims{}
	csrfClaims["user_id"] = userId
	csrfClaims["exp"] = td.CsrfExpires
	csrfClaims["csrf_uuid"] = td.CsrfUuid
	unsignedTK := jwt.NewWithClaims(jwt.SigningMethodHS256, csrfClaims)

	var err error
	td.CsrfToken, err = unsignedTK.SignedString([]byte(csrfSecret))
	if err != nil {
		return nil, customErr.Internal(err.Error())
	}
	return td, nil
}

func (t *tokenOperator) createRefreshToken(userId string, userRole model.Role, td *TokenDetails) (*TokenDetails, error) {

	//Creating Refresh Token
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.BaseUuid + "++" + userId

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userId
	rtClaims["exp"] = td.RtExpires
	rtClaims["user_role"] = userRole

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)

	var err error
	td.RefreshToken, err = rt.SignedString([]byte(refreshSecret))
	if err != nil {
		return nil, customErr.Internal(err.Error())
	}
	return td, nil
}
func (t *tokenOperator) createAccessToken(userId string, userRole model.Role, td *TokenDetails) (*TokenDetails, error) {
	td.AtExpires = time.Now().Add(time.Minute * 30).Unix() //expires after 30 min
	td.BaseUuid = uuid.NewV4().String()
	td.TokenUuid = td.BaseUuid + "@@" + userId

	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.TokenUuid
	atClaims["user_id"] = userId
	atClaims["exp"] = td.AtExpires
	atClaims["user_role"] = userRole
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(accessSecret))
	if err != nil {
		return nil, customErr.Internal(err.Error())
	}
	return td, nil
}

func (t *tokenOperator) CreateStatelessToken(userId string, kind StatelessToken) (string, error) {
	var exp int64

	tokenSecret := ""
	switch kind {
	case ResetToken:
		tokenSecret = passwordResetSecret
		exp = time.Now().Add(time.Minute * 15).Unix() //expires after 15 minutes
	case ValidateUserToken:
		tokenSecret = validationSecret
		exp = time.Now().Add(time.Hour * 24 * 7).Unix() //expires after 7 days
	}

	tkExpires := exp
	tkClaims := jwt.MapClaims{}
	tkClaims["user_id"] = userId
	tkClaims["exp"] = tkExpires

	unsignedTK := jwt.NewWithClaims(jwt.SigningMethodHS256, tkClaims)

	var err error
	token, err := unsignedTK.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", customErr.Internal(err.Error())
	}
	return token, nil
}
func (t *tokenOperator) ExtractStatelessTokenMetadata(ctx context.Context, tokenString string,
	kind StatelessToken) (string, error) {
	tokenSecret := ""
	switch kind {
	case ResetToken:
		tokenSecret = passwordResetSecret
	case ValidateUserToken:
		tokenSecret = validationSecret
	}
	token, err := t.verifyStatelessToken(ctx, tokenString, tokenSecret)
	if err != nil {
		return "", customErr.NoAuth(err.Error())
	}

	userId, err := extractStatelessToken(token)
	if err != nil {
		return "", customErr.NoAuth(err.Error())
	}

	return userId, nil
}

func (t *tokenOperator) verifyStatelessToken(ctx context.Context, resetToken string,
	secret string) (*jwt.Token, error) {
	token, err := t.parse(ctx, resetToken, secret)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func extractStatelessToken(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		userId, userOk := claims["user_id"].(string)
		if !ok || !userOk {
			return "", customErr.NoAuth("unauthorized")

		}
		return userId, nil
	}

	return "", customErr.NoAuth("something went wrong")
}

func (t *tokenOperator) CreateTokens(userId string, userRole model.Role) (*TokenDetails, error) {
	td := &TokenDetails{}

	td, err := t.createAccessToken(userId, userRole, td)
	if err != nil {
		return nil, customErr.Internal(err.Error())
	}
	td, err = t.createRefreshToken(userId, userRole, td)
	if err != nil {
		return nil, customErr.Internal(err.Error())
	}
	td, err = t.createCsrfToken(userId, td)
	if err != nil {
		return nil, customErr.Internal(err.Error())
	}
	return td, nil
}

func (t *tokenOperator) verifyCsrfToken(ctx context.Context, secret string) (*jwt.Token, error) {
	csrfTk, err := t.getCsrfTokenFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	token, err := t.parse(ctx, csrfTk, secret)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (t *tokenOperator) verifyAccessToken(ctx context.Context, secret string) (*jwt.Token, error) {
	tokenMap, err := t.getTokensFromCookie(ctx)
	if err != nil {
		return nil, err
	}
	tokenString := tokenMap["access_token"]
	token, err := t.parse(ctx, tokenString, secret)
	if err != nil {
		return nil, err
	}
	return token, nil
}
func (t *tokenOperator) verifyRefreshToken(ctx context.Context, secret string) (*jwt.Token, error) {
	tokenMap, err := t.getTokensFromCookie(ctx)
	if err != nil {
		return nil, err
	}
	tokenString := tokenMap["refresh_token"]
	token, err := t.parse(ctx, tokenString, secret)
	if err != nil {
		return nil, err
	}
	return token, nil
}
func extractAccessToken(token *jwt.Token, ad *AccessDetails) (*AccessDetails, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		userId, userOk := claims["user_id"].(string)
		role, roleOk := claims["user_role"].(string)
		if !ok || !userOk || !roleOk {
			return nil, customErr.NoAuth("unauthorized")

		}
		ad.TokenUuid = accessUuid
		ad.UserId = userId
		ad.UserRole = model.Role(role)
		return ad, nil
	}

	return nil, customErr.NoAuth("something went wrong")
}

func extractRefreshToken(token *jwt.Token) (*RefreshDetails, error) {
	rd := &RefreshDetails{}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string)
		userId, userOk := claims["user_id"].(string)
		role, roleOk := claims["user_role"].(string)

		if !ok || !userOk || !roleOk {
			return nil, customErr.NoAuth("unauthorized")

		}
		rd.RefreshUuid = refreshUuid
		rd.UserId = userId
		rd.Role = model.Role(role)
		return rd, nil
	}

	return nil, customErr.NoAuth("something went wrong")
}

func extractCsrfToken(token *jwt.Token, ad *AccessDetails) (*AccessDetails, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		csrfUuid, ok := claims["csrf_uuid"].(string)
		userId, userOk := claims["user_id"].(string)
		if !ok || !userOk || userId != ad.UserId {
			return nil, customErr.NoAuth("unauthorized")

		}
		ad.CsrfUuid = csrfUuid
		return ad, nil
	}

	return nil, customErr.NoAuth("something went wrong")
}

func extract(accesToken *jwt.Token, csrfToken *jwt.Token) (*AccessDetails, error) {

	ad := &AccessDetails{}

	ad, err := extractAccessToken(accesToken, ad)
	if err != nil {
		return nil, customErr.NoAuth(err.Error())
	}
	ad, err = extractCsrfToken(csrfToken, ad)
	if err != nil {
		return nil, customErr.NoAuth(err.Error())
	}

	return ad, nil

}

func (t *tokenOperator) ExtractTokenMetadata(ctx context.Context) (*AccessDetails, error) {
	token, err := t.verifyAccessToken(ctx, accessSecret)
	if err != nil {
		return nil, customErr.NoAuth(err.Error())
	}
	csrf, err := t.verifyCsrfToken(ctx, csrfSecret)
	if err != nil {
		return nil, customErr.NoAuth(err.Error())
	}
	acc, err := extract(token, csrf)
	if err != nil {
		return nil, customErr.NoAuth(err.Error())
	}
	return acc, nil
}

func (t *tokenOperator) ExtractRefreshMetadata(ctx context.Context) (*RefreshDetails, error) {
	token, err := t.verifyRefreshToken(ctx, refreshSecret)
	if err != nil {
		return nil, customErr.NoAuth(err.Error())
	}
	rd, err := extractRefreshToken(token)
	if err != nil {
		return nil, customErr.NoAuth(err.Error())
	}
	return rd, nil
}

func (t *tokenOperator) getTokensFromCookie(ctx context.Context) (map[string]string, error) {
	ca, err := middleware.GetCookieAccess(ctx)
	if err != nil {
		return nil, err
	}
	ec := ca.EncodedCookie
	value := make(map[string]string)
	if err = t.sc.Decode("cookie-name", ec, &value); err != nil {
		return nil, customErr.NoAuth(err.Error())
	}

	return value, nil
}

func (t *tokenOperator) getCsrfTokenFromHeader(ctx context.Context) (string, error) {
	ha, err := middleware.GetHeaderAccess(ctx)
	if err != nil {
		return "", err
	}
	csrfTk := ha.CsrfToken
	if len(csrfTk) == 0 {
		return "", customErr.NoAuth("missing csrf token")
	}

	return csrfTk, nil
}

func (t *tokenOperator) parse(ctx context.Context, tokenString string, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, customErr.NoAuth(err.Error())

	}

	return token, nil

}

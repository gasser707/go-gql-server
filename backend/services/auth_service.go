package services

import (
	"context"
	"fmt"
	"os"
	"strconv"

	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graphql/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/gasser707/go-gql-server/middleware"
	"github.com/gasser707/go-gql-server/repo"
	email_svc "github.com/gasser707/go-gql-server/services/email"
	"github.com/gasser707/go-gql-server/utils/auth"
	"github.com/gorilla/securecookie"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
)

var (
	domain = os.Getenv("DOMAIN_NAME")
)

type AuthServiceInterface interface {
	Login(ctx context.Context, input model.LoginInput) (bool, error)
	ValidateCredentials(c context.Context) (IntUserID, model.Role, error)
	Logout(ctx context.Context) (bool, error)
	Refresh(ctx context.Context) (bool, error)
	RequestPasswordReset(ctx context.Context, email string) (bool, error)
	ProcessPasswordReset(ctx context.Context, resetToken string, newPass string) (bool, error)
	ValidateUser(ctx context.Context, validationToken string) (bool, error)
	LogoutAll(ctx context.Context) (bool, error)
}

//UsersService implements the usersServiceInterface
var _ AuthServiceInterface = &authService{}

type authService struct {
	rd           auth.RedisOperatorInterface
	tk           auth.TokenOperatorInterface
	sc           *securecookie.SecureCookie
	repo         repo.AuthRepoInterface
	emailAdaptor email_svc.EmailAdaptorInterface
}

func NewAuthService(db *sqlx.DB, emailAdaptor email_svc.EmailAdaptorInterface) *authService {
	sc := helpers.NewSecureCookie()
	tk := auth.NewTokenOperator(sc)
	rd := auth.NewRedisStore()
	authRepo := repo.NewAuthRepo(db)
	return &authService{rd, tk, sc, authRepo, emailAdaptor}
}

type UserID string
type IntUserID int64

func (s *authService) Login(ctx context.Context, input model.LoginInput) (bool, error) {

	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return false, err
	}

	ok := helpers.CheckPasswordHash(input.Password, user.Password)
	if !ok  {
		return false, customErr.NoAuth("this combination of email password is wrong")
	}

	id := fmt.Sprintf("%v", user.ID)
	role := fmt.Sprintf("%v", user.Role)

	ts, err := s.tk.CreateTokens(id, model.Role(role))
	if err != nil {
		return false, err
	}
	saveErr := s.rd.CreateAuth(id, ts)
	if saveErr != nil {
		return false, err
	}

	ca, err := middleware.GetCookieAccess(ctx)
	if err != nil {
		return false, err
	}
	ca.SetCookie(ts.AccessToken, ts.RefreshToken, s.sc)
	ha, err := middleware.GetHeaderAccess(ctx)
	if err != nil {
		return false, err
	}
	ha.SetCsrfToken(ts.CsrfToken)
	return true, nil
}

func (s *authService) ValidateCredentials(ctx context.Context) (IntUserID, model.Role, error) {
	metadata, err := s.tk.ExtractTokenMetadata(ctx)
	if err != nil {
		return -1, "", err
	}
	userId, err := s.rd.FetchAuth(metadata.TokenUuid, metadata.CsrfUuid)
	if err != nil {
		return -1, "", err
	}

	id, err := strconv.Atoi(userId)
	if err != nil {
		return -1, "", customErr.Internal(err.Error())
	}

	return IntUserID(id), metadata.UserRole, nil
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

func (s *authService) RequestPasswordReset(ctx context.Context, email string) (bool, error) {

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return true, nil
	}

	token, err := s.tk.CreateStatelessToken(fmt.Sprintf("%d", user.ID), auth.ResetToken)
	if err != nil {
		return false, err
	}

	go s.emailAdaptor.SendResetPassEmail(ctx, "auth@shotify.com", []string{email},
		user.Username, fmt.Sprintf("http://%s/reset?token=%s", domain, token))

	return true, nil

}

func (s *authService) ProcessPasswordReset(ctx context.Context, resetToken string, newPass string) (bool, error) {
	//If metadata is passed and the tokens valid, delete them from the redis store
	userId, err := s.tk.ExtractStatelessTokenMetadata(ctx, resetToken, auth.ResetToken)
	if err != nil {
		return false, err
	}
	hashedPwd, err := helpers.HashPassword(newPass)
	if err != nil {
		return false, err
	}
	err = s.repo.UpdatePassword(ctx, userId, hashedPwd)
	if err != nil {
		return true, nil
	}
	return true, nil

}

func (s *authService) ValidateUser(ctx context.Context, validationToken string) (bool, error) {
	//If metadata is passed and the tokens valid, delete them from the redis store
	userId, err := s.tk.ExtractStatelessTokenMetadata(ctx, validationToken, auth.ValidateUserToken)
	if err != nil {
		return false, err
	}
	 err = s.repo.UpdateVerified(ctx, userId)
	if err != nil {
		return true, nil
	}

	return true, nil

}

func (s *authService) LogoutAll(ctx context.Context) (bool, error) {
	metadata, err := s.tk.ExtractTokenMetadata(ctx)
	if err != nil {
		return false, err
	}
	err = s.rd.DeleteAllUserTokens(metadata.UserId)
	if err != nil {
		return false, err
	}
	return true, nil

}

func (s *authService) Refresh(ctx context.Context) (bool, error) {
	metadata, err := s.tk.ExtractRefreshMetadata(ctx)
	if err != nil {
		return false, err
	}
	userId, err := s.rd.FetchRefresh(metadata.RefreshUuid)
	if err != nil {
		return false, err
	}

	//Delete the previous Refresh Token
	delErr := s.rd.DeleteRefresh(metadata.RefreshUuid)
	if delErr != nil { //if any goes wrong
		return false, err
	}
	//Create new pairs of refresh and access csrf tokens
	ts, createErr := s.tk.CreateTokens(userId, model.Role(metadata.Role))
	if createErr != nil {
		return false, err
	}
	//save the tokens metadata to redis
	saveErr := s.rd.CreateAuth(userId, ts)
	if saveErr != nil {
		return false, err
	}
	return true, nil
}

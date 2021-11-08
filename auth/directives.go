package auth

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gasser707/go-gql-server/graph/model"
)

 var Authorize = func (ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	  _, _, err:= AuthService.validateCredentials(ctx)

	  if(err!=nil){
	  return nil, fmt.Errorf("no auth credentials found")
	  }

      return next(ctx)
}

var AdminOrMod = func (ctx context.Context, obj model.LoginInput, next graphql.Resolver) (interface{}, error) {
	_, role, err:= AuthService.validateCredentials(ctx)
	if(err!=nil){
		return nil, fmt.Errorf("no auth credentials found")
	}
	if role != model.RoleAdmin && role != model.RoleModerator {
		return nil, fmt.Errorf("unauthorized")

	}
	return next(ctx)
}

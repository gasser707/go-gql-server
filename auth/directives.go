package auth

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gasser707/go-gql-server/graph/model"
)

 var Authorize = func (ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	  _, err:= AuthService.GetCredentials(ctx)

	  if(err!=nil){
	  return nil, fmt.Errorf("no auth credentials found")
	  }

      return next(ctx)
}

var AdminOrSelf = func (ctx context.Context, obj model.LoginInput, next graphql.Resolver) (interface{}, error) {
	_, err:= AuthService.GetCredentials(ctx)
	if(err!=nil){
		return nil, fmt.Errorf("no auth credentials found")
	}
	// uId = md.UserId
	return next(ctx)
}

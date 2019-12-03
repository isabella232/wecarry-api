package actions

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/gqlgen"
	"github.com/vektah/gqlparser/gqlerror"
)

func GQLHandler(c buffalo.Context) error {
	gqlSuccess := true

	h := handler.GraphQL(gqlgen.NewExecutableSchema(gqlgen.Config{Resolvers: &gqlgen.Resolver{}}),
		handler.ErrorPresenter(
			func(ctx context.Context, e error) *gqlerror.Error {
				gqlSuccess = false
				return graphql.DefaultErrorPresenter(ctx, e)
			}))

	newCtx := context.WithValue(c.Request().Context(), "BuffaloContext", c)
	h.ServeHTTP(c.Response(), c.Request().WithContext(newCtx))

	if !gqlSuccess {
		return nil
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		domain.ErrLogger.Printf("no transaction found in GQLHandler")
	}
	if err := tx.TX.Commit(); err != nil {
		domain.ErrLogger.Printf("database commit failed, %s", err)
	}
	return nil
}

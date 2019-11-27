package actions

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/handler"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/gqlgen"
)

func GQLHandler(c buffalo.Context) error {
	h := handler.GraphQL(gqlgen.NewExecutableSchema(gqlgen.Config{Resolvers: &gqlgen.Resolver{}}))
	newCtx := context.WithValue(c.Request().Context(), "BuffaloContext", c)
	h.ServeHTTP(c.Response(), c.Request().WithContext(newCtx))

	if res, ok := c.Response().(*buffalo.Response); ok {
		if res.Status < 200 || res.Status >= 400 {
			return errors.New("non-success error code returned from gqlgen handler")
		}
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.New("no transaction found")
	}
	if err := tx.TX.Commit(); err != nil {
		domain.ErrLogger.Printf("database commit failed, %s", err)
		return err
	}
	return nil
}

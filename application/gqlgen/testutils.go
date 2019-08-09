package gqlgen

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/ast"
)

func BounceTestDB() error {
	test, err := pop.Connect("test")
	if err != nil {
		return err
	}

	// drop the test db:
	test.Dialect.DropDB()

	// create the test db:
	err = test.Dialect.CreateDB()
	if err != nil {
		return err
	}

	fm, err := pop.NewFileMigrator("../migrations", test)
	if err != nil {
		return err
	}

	if err := fm.Up(); err != nil {
		return err
	}

	return nil
}

func createOrgs(fixtures models.Organizations) error {
	for _, f := range fixtures {
		if err := models.DB.Create(&f); err != nil {
			return fmt.Errorf("error creating org %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func createUsers(fixtures models.Users) error {
	for _, f := range fixtures {
		if err := models.DB.Create(&f); err != nil {
			return fmt.Errorf("error creating user %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func createUserOrgs(fixtures models.UserOrganizations) error {
	for _, f := range fixtures {
		if err := models.DB.Create(&f); err != nil {
			return fmt.Errorf("error creating user-org %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func createPosts(fixtures models.Posts) error {
	for _, f := range fixtures {
		if err := models.DB.Create(&f); err != nil {
			return fmt.Errorf("error creating post %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func createThreads(fixtures models.Threads) error {
	db := models.DB
	for _, f := range fixtures {
		if err := db.Create(&f); err != nil {
			return fmt.Errorf("error creating thread %+v ...\n %v \n", f, err)
		}
	}

	threads := []models.Thread{}
	if err := db.All(&threads); err != nil {
		return fmt.Errorf("error retrieving new threads ... %v \n", err)
	}

	if len(threads) < len(fixtures) {
		return fmt.Errorf("wrong number of threads created, expected %v, but got %v", len(fixtures), len(threads))
	}

	return nil
}

func createThreadParticipants(fixtures models.ThreadParticipants) error {
	for _, f := range fixtures {
		if err := models.DB.Create(&f); err != nil {
			return fmt.Errorf("error creating threadparticipant %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func createMessages(fixtures models.Messages) error {
	for _, f := range fixtures {
		if err := models.DB.Create(&f); err != nil {
			return fmt.Errorf("error creating message %+v ...\n %v \n", f, err)
		}
	}
	return nil
}

func testContext(sel ast.SelectionSet) context.Context {

	ctx := context.Background()

	rqCtx := &graphql.RequestContext{}
	ctx = graphql.WithRequestContext(ctx, rqCtx)

	root := &graphql.ResolverContext{
		Field: graphql.CollectedField{
			Selections: sel,
		},
	}
	ctx = graphql.WithResolverContext(ctx, root)

	return ctx
}
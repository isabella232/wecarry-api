package gqlgen

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
)

var PostRoleMap = map[PostRole]string{
	PostRoleCreatedby: models.PostRoleCreatedby,
	PostRoleReceiving: models.PostRoleReceiving,
	PostRoleProviding: models.PostRoleProviding,
}

// UserFields maps GraphQL fields to their equivalent database fields. For related types, the
// foreign key field name is provided.
func UserFields() map[string]string {
	return map[string]string{
		"id":          "uuid",
		"email":       "email",
		"firstName":   "first_name",
		"lastName":    "last_name",
		"nickname":    "nickname",
		"accessToken": "access_token",
		"createdAt":   "created_at",
		"updatedAt":   "updated_at",
		"adminRole":   "admin_role",
		"photoURL":    "photo_url",
		"photoFile":   "photo_file_id",
	}
}

func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

func (r *userResolver) ID(ctx context.Context, obj *models.User) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

func (r *userResolver) AdminRole(ctx context.Context, obj *models.User) (*Role, error) {
	if obj == nil {
		return nil, nil
	}
	a := Role(obj.AdminRole.String)
	return &a, nil
}

func (r *userResolver) Organizations(ctx context.Context, obj *models.User) ([]*models.Organization, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetOrganizations()
}

func (r *userResolver) Posts(ctx context.Context, obj *models.User, role PostRole) ([]*models.Post, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetPosts(PostRoleMap[role])
}

func (r *userResolver) PhotoURL(ctx context.Context, obj *models.User) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.GetPhotoURL()
}

func (r *queryResolver) Users(ctx context.Context) ([]*models.User, error) {
	db := models.DB
	var dbUsers []*models.User

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin {
		err := fmt.Errorf("not authorized")
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return []*models.User{}, err
	}

	if err := db.Select(GetSelectFieldsForUsers(ctx)...).All(&dbUsers); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting users: %v", err.Error()))
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return []*models.User{}, err
	}

	return dbUsers, nil
}

func (r *queryResolver) User(ctx context.Context, id *string) (*models.User, error) {
	dbUser := models.User{}

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if id == nil {
		return &currentUser, nil
	}

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin && currentUser.Uuid.String() != *id {
		err := fmt.Errorf("not authorized")
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &dbUser, err
	}

	if err := models.DB.Where("uuid = ?", id).First(&dbUser); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting user: %v", err.Error()))
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &dbUser, err
	}

	return &dbUser, nil
}

func GetSelectFieldsForUsers(ctx context.Context) []string {
	selectFields := GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx))
	selectFields = append(selectFields, "id")
	if domain.IsStringInSlice("photoURL", graphql.CollectAllFields(ctx)) {
		selectFields = append(selectFields, "photo_file_id")
	}
	return selectFields
}

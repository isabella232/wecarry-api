package gqlgen

import (
	"context"
	"errors"
	"strings"

	"github.com/silinternational/wecarry-api/models"
)

// postRoleMap is used to convert PostRole gql enum values to values used by models
var postRoleMap = map[PostRole]string{
	PostRoleCreatedby: models.PostsCreated,
	PostRoleReceiving: models.PostsReceiving,
	PostRoleProviding: models.PostsProviding,
}

// User is required by gqlgen
func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

// ID provides the UUID instead of the autoincrement ID.
func (r *userResolver) ID(ctx context.Context, obj *models.User) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}

// Organizations retrieves the list of Organizations to which the queried user is associated
func (r *userResolver) Organizations(ctx context.Context, obj *models.User) ([]models.Organization, error) {
	if obj == nil {
		return nil, nil
	}

	organizations, err := obj.GetOrganizations()
	if err != nil {
		return nil, reportError(ctx, err, "GetUserOrganizations")
	}

	return organizations, nil
}

// Posts retrieves the list of Posts associated with the queried user, where association is defined by the given `role`.
func (r *userResolver) Posts(ctx context.Context, obj *models.User, role PostRole) ([]models.Post, error) {
	if obj == nil {
		return nil, nil
	}

	posts, err := obj.GetPosts(postRoleMap[role])
	if err != nil {
		extras := map[string]interface{}{
			"role": role,
		}
		return nil, reportError(ctx, err, "GetUserPosts", extras)
	}

	return posts, nil
}

// AvatarURL retrieves a URL for the user profile photo or avatar.
func (r *userResolver) AvatarURL(ctx context.Context, obj *models.User) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	photoURL, err := obj.GetPhotoURL()
	if err != nil {
		return nil, reportError(ctx, err, "GetUserPhotoURL")
	}

	return photoURL, nil
}

// Location retrieves the queried user's location.
func (r *userResolver) Location(ctx context.Context, obj *models.User) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}

	location, err := obj.GetLocation()
	if err != nil {
		return nil, reportError(ctx, err, "GetUserLocation")
	}

	return location, nil
}

// UnreadMessageCount calculates the number of unread messages for the queried user
func (r *userResolver) UnreadMessageCount(ctx context.Context, obj *models.User) (int, error) {
	if obj == nil {
		return 0, nil
	}
	mCounts, err := obj.UnreadMessageCount()

	if err != nil {
		return 0, reportError(ctx, err, "GetUserUnreadMessageCount")
	}
	total := 0
	for _, c := range mCounts {
		total += c.Count
	}

	return total, nil
}

// Users retrieves a list of users
func (r *queryResolver) Users(ctx context.Context) ([]models.User, error) {
	currentUser := models.GetCurrentUserFromGqlContext(ctx)

	role := currentUser.AdminRole
	if role != models.UserAdminRoleSuperAdmin {
		err := errors.New("insufficient permissions")
		extras := map[string]interface{}{
			"role": role,
		}
		return nil, reportError(ctx, err, "GetUsers.Unauthorized", extras)
	}

	users := models.Users{}
	if err := users.All(); err != nil {
		return nil, reportError(ctx, err, "GetUsers")
	}

	return users, nil
}

// User retrieves a single user
func (r *queryResolver) User(ctx context.Context, id *string) (*models.User, error) {
	currentUser := models.GetCurrentUserFromGqlContext(ctx)

	if id == nil {
		return &currentUser, nil
	}

	role := currentUser.AdminRole
	if role != models.UserAdminRoleSuperAdmin && currentUser.UUID.String() != *id {
		err := errors.New("insufficient permissions")
		extras := map[string]interface{}{
			"role": role,
		}
		return nil, reportError(ctx, err, "GetUser.Unauthorized", extras)
	}

	dbUser := models.User{}
	if err := dbUser.FindByUUID(*id); err != nil {
		return nil, reportError(ctx, err, "GetUser")
	}

	return &dbUser, nil
}

// UpdateUser takes data from the GraphQL `UpdateUser` mutation and updates the database. If the
// user ID is provided and the current user is allowed to edit profiles, that user will be updated.
// Otherwise, the current authenticated user is updated.
func (r *mutationResolver) UpdateUser(ctx context.Context, input UpdateUserInput) (*models.User, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	var user models.User

	if input.ID != nil {
		if err := user.FindByUUID(*(input.ID)); err != nil {
			return nil, reportError(ctx, err, "UpdateUser.NotFound")
		}
	} else {
		user = cUser
	}

	if cUser.AdminRole != models.UserAdminRoleSuperAdmin && cUser.ID != user.ID {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "UpdateUser.Unauthorized")
	}

	if input.Nickname != nil {
		user.Nickname = *input.Nickname
	}

	if input.PhotoID != nil {
		if _, err := user.AttachPhoto(*input.PhotoID); err != nil {
			return nil, reportError(ctx, err, "UpdateUser.PhotoNotFound")
		}
	}

	if input.Location != nil {
		if err := user.SetLocation(convertGqlLocationInputToDBLocation(*input.Location)); err != nil {
			return nil, reportError(ctx, err, "UpdateUser.SetLocationError")
		}
	}

	// No deleting of preferences supported at this time
	if input.Preferences != nil {
		standardPrefs, err := convertUserPreferencesToStandardPreferences(input.Preferences)

		if err != nil {
			return nil, reportError(ctx, err, "UpdateUser.PreferencesInput")
		}

		if _, err = user.UpdateStandardPreferences(standardPrefs); err != nil {
			return nil, reportError(ctx, err, "UpdateUser.Preferences")
		}
	}

	if err := user.Save(); err != nil {
		return nil, reportError(ctx, err, "UpdateUser")
	}

	return &user, nil
}

// Preferences resolves the `preferences` property of the user query, retrieving the related records from the database
// and using them to hydrate a StandardPreferences struct.
func (r *userResolver) Preferences(ctx context.Context, obj *models.User) (*models.StandardPreferences, error) {
	if obj == nil {
		return nil, nil
	}

	user := models.GetCurrentUserFromGqlContext(ctx)
	standardPrefs, err := obj.GetPreferences()
	if err != nil {
		extras := map[string]interface{}{
			"user": user.UUID,
		}
		return nil, reportError(ctx, err, "GetUserPreferences", extras)
	}

	// These have particular acceptable values, unlike TimeZone
	standardPrefs.Language = strings.ToUpper(standardPrefs.Language)
	standardPrefs.WeightUnit = strings.ToUpper(standardPrefs.WeightUnit)

	return &standardPrefs, nil
}

// getPublicProfiles converts a list of models.User to PublicProfile, hiding private profile information
func getPublicProfiles(ctx context.Context, users []models.User) []PublicProfile {
	profiles := make([]PublicProfile, len(users))
	for i, p := range users {
		prof := getPublicProfile(ctx, &p)
		profiles[i] = *prof
	}
	return profiles
}

// getPublicProfile converts a models.User to a PublicProfile, which hides private profile information
func getPublicProfile(ctx context.Context, user *models.User) *PublicProfile {
	if user == nil {
		return nil
	}

	url, err := user.GetPhotoURL()
	if err != nil {
		_ = reportError(ctx, err, "", map[string]interface{}{"user": user.UUID})
		return nil
	}

	return &PublicProfile{
		ID:        user.UUID.String(),
		Nickname:  user.Nickname,
		AvatarURL: url,
	}
}

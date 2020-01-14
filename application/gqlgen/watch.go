package gqlgen

import (
	"context"
	"errors"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/models"
)

// Watch returns the watch resolver. It is required by GraphQL
func (r *Resolver) Watch() WatchResolver {
	return &watchResolver{r}
}

type watchResolver struct{ *Resolver }

// ID resolves the `ID` property of the watch query. It provides the UUID instead of the autoincrement ID.
func (r *watchResolver) ID(ctx context.Context, obj *models.Watch) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}

// Owner resolves the `owner` property of the watch query. It retrieves the related record from the database.
func (r *watchResolver) Owner(ctx context.Context, obj *models.Watch) (*PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	creator, err := obj.GetOwner()
	if err != nil {
		return nil, reportError(ctx, err, "GetWatchCreator")
	}

	return getPublicProfile(ctx, creator), nil
}

// Location resolves the `location` property of the watch query, retrieving the related record from the database.
func (r *watchResolver) Location(ctx context.Context, obj *models.Watch) (*models.Location, error) {
	if obj == nil {
		return &models.Location{}, nil
	}

	location, err := obj.GetLocation()
	if err != nil {
		return &models.Location{}, reportError(ctx, err, "GetWatchLocation")
	}

	return location, nil
}

// MyWatches resolves the `myWatches` query by getting a list of Watches owned by the current user
func (r *queryResolver) MyWatches(ctx context.Context) ([]models.Watch, error) {
	watches := models.Watches{}
	currentUser := models.GetCurrentUserFromGqlContext(ctx)
	if err := watches.FindByUser(currentUser); err != nil {
		extras := map[string]interface{}{
			"user": currentUser.UUID,
		}
		return nil, reportError(ctx, err, "MyWatches", extras)
	}

	return watches, nil
}

// convertWatchInput takes a `WatchInput` and either finds a record matching the UUID given in `input.ID` or
// creates a new `models.Watch` with a new UUID. In either case, all properties that are not `nil` are set to the value
// provided in `input`
func convertWatchInput(ctx context.Context, input watchInput, currentUser models.User) (models.Watch, error) {
	watch := models.Watch{}

	if input.ID != nil {
		if err := watch.FindByUUID(*input.ID); err != nil {
			return watch, err
		}
	} else {
		watch.OwnerID = currentUser.ID
	}

	return watch, nil
}

type watchInput struct {
	ID       *string
	Location *LocationInput
}

// CreateWatch resolves the `createWatch` mutation.
func (r *mutationResolver) CreateWatch(ctx context.Context, input watchInput) (*models.Watch, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	watch, err := convertWatchInput(ctx, input, cUser)
	if err != nil {
		return nil, reportError(ctx, err, "CreateWatch.ProcessInput", extras)
	}

	location := convertGqlLocationInputToDBLocation(*input.Location)
	if err = location.Create(); err != nil {
		return nil, reportError(ctx, err, "CreateWatch.SetLocation", extras)
	}
	watch.LocationID = nulls.NewInt(location.ID)

	if err = watch.Create(); err != nil {
		return nil, reportError(ctx, err, "CreateWatch", extras)
	}

	return &watch, nil
}

// UpdateWatch resolves the `updateWatch` mutation.
func (r *mutationResolver) UpdateWatch(ctx context.Context, input watchInput) (*models.Watch, error) {
	currentUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": currentUser.UUID,
	}

	watch, err := convertWatchInput(ctx, input, currentUser)
	if err != nil {
		return nil, reportError(ctx, err, "UpdateWatch.ProcessInput", extras)
	}

	if watch.OwnerID != currentUser.ID {
		return nil, reportError(ctx, errors.New("user attempted to update non-owned Watch"),
			"UpdateWatch.NotFound", extras)
	}

	if err := watch.Update(); err != nil {
		return nil, reportError(ctx, err, "UpdateWatch", extras)
	}

	if input.Location != nil {
		if err = watch.SetLocation(convertGqlLocationInputToDBLocation(*input.Location)); err != nil {
			return nil, reportError(ctx, err, "UpdateWatch.SetLocation", extras)
		}
	}

	return &watch, nil
}

// RemoveWatch resolves the `removeWatch` mutation.
func (r *mutationResolver) RemoveWatch(ctx context.Context, input RemoveWatchInput) ([]models.Watch, error) {
	currentUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": currentUser.UUID,
	}

	var watch models.Watch
	if err := watch.FindByUUID(input.ID); err != nil {
		return nil, reportError(ctx, err, "RemoveWatch.NotFound", extras)
	}

	if watch.OwnerID != currentUser.ID {
		return nil, reportError(ctx, errors.New("user attempted to delete non-owned Watch"),
			"RemoveWatch.NotFound", extras)
	}

	if err := watch.Destroy(); err != nil {
		return nil, reportError(ctx, err, "RemoveWatch", extras)
	}

	var watches models.Watches
	if err := watches.FindByUser(currentUser); err != nil {
		return nil, reportError(ctx, err, "RemoveWatch.FindByUser", extras)
	}

	return watches, nil
}
package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
)

// Watch is the model for storing request watches that trigger notifications on the conditions specified
type Watch struct {
	ID            int          `json:"id" db:"id"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
	UUID          uuid.UUID    `json:"uuid" db:"uuid"`
	OwnerID       int          `json:"owner_id" db:"owner_id"`
	Name          string       `json:"name" db:"name"`
	DestinationID nulls.Int    `json:"destination_id" db:"destination_id"`
	OriginID      nulls.Int    `json:"origin_id" db:"origin_id"`
	MeetingID     nulls.Int    `json:"meeting_id" db:"meeting_id"`
	SearchText    nulls.String `json:"search_text" db:"search_text"`
	Size          *RequestSize `json:"size" db:"size"`
}

// Watches is used for methods that operate on lists of objects
type Watches []Watch

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (w *Watch) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: w.UUID, Name: "UUID"},
		&validators.IntIsPresent{Field: w.OwnerID, Name: "OwnerID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (w *Watch) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (w *Watch) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Create stores the Watch data as a new record in the database.
func (w *Watch) Create() error {
	return create(w)
}

// Update writes the Watch data to an existing database record.
func (w *Watch) Update() error {
	return update(w)
}

// FindByUUID loads from DB the Watch record identified by the given UUID
func (w *Watch) FindByUUID(id string) error {
	if id == "" {
		return errors.New("error: watch uuid must not be blank")
	}

	if err := DB.Where("uuid = ?", id).First(w); err != nil {
		return fmt.Errorf("error finding watch by uuid: %s", err.Error())
	}

	return nil
}

// FindByUser returns all watches owned by the given user.
func (w *Watches) FindByUser(user User) error {
	if err := DB.Where("owner_id = ?", user.ID).Order("updated_at desc").All(w); err != nil {
		return err
	}

	return nil
}

// GetOwner returns the owner of the watch.
func (w *Watch) GetOwner() (*User, error) {
	owner := User{}
	if err := DB.Find(&owner, w.OwnerID); err != nil {
		return nil, err
	}
	return &owner, nil
}

// GetDestination does not check authorization
func (w *Watch) GetDestination() (*Location, error) {
	location := &Location{}
	if !w.DestinationID.Valid {
		return nil, nil
	}
	if err := DB.Find(location, w.DestinationID); err != nil {
		return nil, err
	}
	return location, nil
}

// GetOrigin does not check authorization
func (w *Watch) GetOrigin() (*Location, error) {
	location := &Location{}
	if !w.OriginID.Valid {
		return nil, nil
	}
	if err := DB.Find(location, w.OriginID); err != nil {
		return nil, err
	}
	return location, nil
}

// SetDestination sets the destination field, creating a new record in the database if necessary.
func (w *Watch) SetDestination(location Location) error {
	if w.DestinationID.Valid {
		location.ID = w.DestinationID.Int
		return location.Update()
	}
	if err := location.Create(); err != nil {
		return err
	}
	w.DestinationID = nulls.NewInt(location.ID)
	return w.Update()
}

// SetOrigin sets the origin field, creating a new record in the database if necessary.
func (w *Watch) SetOrigin(location Location) error {
	if w.OriginID.Valid {
		location.ID = w.OriginID.Int
		return location.Update()
	}
	if err := location.Create(); err != nil {
		return err
	}
	w.OriginID = nulls.NewInt(location.ID)
	return w.Update()
}

// Destroy wraps the Pop function of the same name
func (w *Watch) Destroy() error {
	return DB.Destroy(w)
}

func (w *Watch) Meeting(ctx context.Context) (*Meeting, error) {
	if w == nil || !w.MeetingID.Valid || CurrentUser(ctx).ID != w.OwnerID {
		return nil, nil
	}
	meeting := &Meeting{}
	if err := DB.Find(meeting, w.MeetingID); err != nil {
		return nil, err
	}
	return meeting, nil
}

// matchesRequest returns true if all non-null watch criteria match the request
func (w *Watch) matchesRequest(request Request) bool {
	if w == nil {
		domain.ErrLogger.Printf("nil receiver in Watch.matchesRequest")
		return false
	}
	matchFunctions := []func(*Watch, Request) bool{
		(*Watch).sizeMatches,
		(*Watch).textMatches,
		(*Watch).meetingMatches,
		(*Watch).destinationMatches,
		(*Watch).originMatches,
	}
	for _, c := range matchFunctions {
		if !c(w, request) {
			return false
		}
	}
	return true
}

// destinationMatches returns true if watch destination is not provided or passes the IsNear test
func (w *Watch) destinationMatches(request Request) bool {
	if w == nil {
		domain.ErrLogger.Printf("nil receiver in Watch.destinationMatches")
		return false
	}
	if !w.DestinationID.Valid {
		return true
	}
	requestDestination, err := request.GetDestination()
	if err != nil {
		domain.ErrLogger.Printf("failed to get request %s destination in destinationMatches, %s", request.UUID, err)
		return false
	}
	watchDestination, err := w.GetDestination()
	if err != nil {
		domain.ErrLogger.Printf("failed to get watch %s destination in destinationMatches, %s", w.UUID, err)
	}
	return watchDestination.IsNear(*requestDestination)
}

// originMatches returns true if watch origin is not provided or passes the IsNear test
func (w *Watch) originMatches(request Request) bool {
	if w == nil {
		domain.ErrLogger.Printf("nil receiver in Watch.originMatches")
		return false
	}
	if !w.OriginID.Valid {
		return true
	}
	requestOrigin, err := request.GetOrigin()
	if err != nil {
		domain.ErrLogger.Printf("failed to get request %s origin in originMatches, %s", request.UUID, err)
		return false
	}
	watchOrigin, err := w.GetOrigin()
	if err != nil {
		domain.ErrLogger.Printf("failed to get watch %s origin in originMatches, %s", w.UUID, err)
	}
	return watchOrigin.IsNear(*requestOrigin)
}

// meetingMatches returns true if watch meeting is not provided or is identical to the request meeting
func (w *Watch) meetingMatches(request Request) bool {
	if w == nil {
		domain.ErrLogger.Printf("nil receiver in Watch.meetingMatches")
		return false
	}
	if !w.MeetingID.Valid {
		return true
	}
	return w.MeetingID == request.MeetingID
}

// textMatches returns true if watch text is not provided or is in the request title, description or creator's nickname
func (w *Watch) textMatches(request Request) bool {
	if w == nil {
		domain.ErrLogger.Printf("nil receiver in Watch.textMatches")
		return false
	}
	if !w.SearchText.Valid {
		return true
	}
	if strings.Contains(request.Title, w.SearchText.String) {
		return true
	}
	if strings.Contains(request.Description.String, w.SearchText.String) {
		return true
	}
	creator, err := request.Creator()
	if err != nil {
		domain.ErrLogger.Printf("failed to get request %s creator in textMatches, %s", request.UUID, err)
		return false
	}
	if strings.Contains(creator.Nickname, w.SearchText.String) {
		return true
	}
	return false
}

// sizeMatches returns true if watch size is larger or the same as the request size
func (w *Watch) sizeMatches(request Request) bool {
	if w == nil {
		domain.ErrLogger.Printf("nil receiver in Watch.sizeMatches")
		return false
	}
	if w.Size == nil {
		return true
	}
	return w.Size.isLargerOrSame(request.Size)
}

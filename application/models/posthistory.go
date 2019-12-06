package models

import (
	"encoding/json"
	"fmt"
	"github.com/silinternational/wecarry-api/domain"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
)

type PostHistory struct {
	ID         int        `json:"id" db:"id"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	Status     PostStatus `json:"status" db:"status"`
	PostID     int        `json:"post_id" db:"post_id"`
	ReceiverID nulls.Int  `json:"receiver_id" db:"receiver_id"`
	ProviderID nulls.Int  `json:"provider_id" db:"provider_id"`
	Post       Post       `belongs_to:"posts"`
	Receiver   User       `belongs_to:"users"`
	Provider   User       `belongs_to:"users"`
}

// String is not required by pop and may be deleted
func (p PostHistory) String() string {
	jt, _ := json.Marshal(p)
	return string(jt)
}

// PostHistories is not required by pop and may be deleted
type PostHistories []PostHistory

// String is not required by pop and may be deleted
func (p PostHistories) String() string {
	jt, _ := json.Marshal(p)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (p *PostHistory) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (p *PostHistory) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (p *PostHistory) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Load reads the selected fields from the database
func (p *PostHistory) Load(fields ...string) error {
	if err := DB.Load(p, fields...); err != nil {
		return fmt.Errorf("error loading data for post history %v, %s", p.ID, err)
	}

	return nil
}

// createForPost checks if the post has a status that is different than the
// most recent of its Post History entries.  If so, it creates a new Post History
// with the Post's new status.
func (pH PostHistory) createForPost(post Post) error {
	err := DB.Where("post_id = ?", post.ID).Last(&pH)

	if domain.IsOtherThanNoRows(err) {
		return err
	}

	if pH.Status != post.Status {
		newPH := PostHistory{
			Status:     post.Status,
			PostID:     post.ID,
			ReceiverID: post.ReceiverID,
			ProviderID: post.ProviderID,
		}

		if err := DB.Create(&newPH); err != nil {
			return err
		}
	}

	return nil
}

// pop deletes the most recent postHistory entry for a post
// assuming it's status matches the expected one.
func (pH PostHistory) popForPost(post Post, currentStatus PostStatus) error {
	if err := DB.Where("post_id = ?", post.ID).Last(&pH); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return err
		}
		domain.ErrLogger.Printf(
			"error popping post histories for post id %v. None Found", post.ID)
		return nil
	}

	if pH.Status != currentStatus {
		domain.ErrLogger.Printf(
			"error popping post histories for post id %v. Expected newStatus %s but found %s",
			post.ID, currentStatus, pH.Status)
		return nil
	}

	if err := DB.Destroy(&pH); err != nil {
		return err
	}

	return nil
}
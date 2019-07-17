package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/envy"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type User struct {
	ID           int               `json:"id" db:"id"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`
	Email        string            `json:"email" db:"email"`
	FirstName    string            `json:"first_name" db:"first_name"`
	LastName     string            `json:"last_name" db:"last_name"`
	Nickname     string            `json:"nickname" db:"nickname"`
	AuthOrgID    int               `json:"auth_org_id" db:"auth_org_id"`
	AuthOrgUid   string            `json:"auth_org_uid" db:"auth_org_uid"`
	AdminRole    nulls.String      `json:"admin_role" db:"admin_role"`
	Uuid         string            `json:"uuid" db:"uuid"`
	AuthOrg      Organization      `belongs_to:"organizations"`
	AccessTokens []UserAccessToken `has_many:"user_access_tokens"`
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: u.ID, Name: "ID"},
		&validators.StringIsPresent{Field: u.Email, Name: "Email"},
		&validators.StringIsPresent{Field: u.FirstName, Name: "FirstName"},
		&validators.StringIsPresent{Field: u.LastName, Name: "LastName"},
		&validators.StringIsPresent{Field: u.Nickname, Name: "Nickname"},
		&validators.IntIsPresent{Field: u.AuthOrgID, Name: "AuthOrgID"},
		&validators.StringIsPresent{Field: u.AuthOrgUid, Name: "AuthOrgUid"},
		&validators.StringIsPresent{Field: u.Uuid, Name: "Uuid"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// CreateAccessToken - Create and store new UserAccessToken
func (u *User) CreateAccessToken(tx *pop.Connection, clientID string) (string, int64, error) {

	token := createAccessTokenPart()
	hash := hashClientIdAccessToken(clientID + token)
	expireAt := createAccessTokenExpiry()

	userAccessToken := &UserAccessToken{
		UserID:      u.ID,
		AccessToken: hash,
		ExpiresAt:   expireAt,
	}

	err := tx.Save(userAccessToken)
	if err != nil {
		return "", 0, err
	}

	return token, expireAt.UTC().Unix(), nil
}

func FindUserByAccessToken(accessToken string) (User, error) {

	userAccessToken := UserAccessToken{}

	if accessToken == "" {
		return User{}, fmt.Errorf("error: access token must not be blank")
	}

	dbAccessToken := hashClientIdAccessToken(accessToken)
	queryString := fmt.Sprintf("access_token = '%s'", dbAccessToken)

	if err := DB.Eager().Where(queryString).First(&userAccessToken); err != nil {
		return User{}, fmt.Errorf("error finding user by access token: %s", err.Error())
	}

	if userAccessToken.ExpiresAt.Before(time.Now()) {
		err := DB.Destroy(userAccessToken)
		if err != nil {
			log.Printf("Unable to delete expired userAccessToken, id: %v", userAccessToken.ID)
		}
		return User{}, fmt.Errorf("access token has expired")
	}

	return userAccessToken.User, nil
}

func createAccessTokenExpiry() time.Time {
	lifetime := envy.Get("ACCESS_TOKEN_LIFETIME", "28800")

	lifetimeSeconds, err := strconv.Atoi(lifetime)
	if err != nil {
		lifetimeSeconds = 28800
	}

	dtNow := time.Now()
	futureTime := dtNow.Add(time.Second * time.Duration(lifetimeSeconds))

	return futureTime
}

func createAccessTokenPart() string {
	var alphanumerics = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	tokenLength := 32
	b := make([]rune, tokenLength)
	for i := range b {
		b[i] = alphanumerics[rand.Intn(len(alphanumerics))]
	}

	accessToken := string(b)

	return accessToken
}

func hashClientIdAccessToken(accessToken string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(accessToken)))
}

package models

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/silinternational/handcarry-api/auth"

	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/handcarry-api/domain"
)

func (ms *ModelSuite) TestUser_FindOrCreateFromAuthUser() {
	t := ms.T()
	resetTables(t)

	// create org for test
	org := &Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := models.DB.Create(org)
	if err != nil {
		t.Errorf("Failed to create organization for test, error: %s", err)
		t.FailNow()
	}

	type args struct {
		tx       *pop.Connection
		orgID    int
		authUser *auth.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantID  int
	}{
		{
			name: "create new user: test_user1",
			args: args{
				orgID: org.ID,
				authUser: &auth.User{
					FirstName: "Test",
					LastName:  "User",
					Email:     "test_user1@domain.com",
					UserID:    "test_user1",
				},
			},
			wantErr: false,
			wantID:  1,
		},
		{
			name: "find existing user: test_user1",
			args: args{
				orgID: org.ID,
				authUser: &auth.User{
					FirstName: "Test",
					LastName:  "User",
					Email:     "test_user1@domain.com",
					UserID:    "test_user1",
				},
			},
			wantErr: false,
			wantID:  1,
		},
		{
			name: "conflicting user",
			args: args{
				orgID: org.ID,
				authUser: &auth.User{
					FirstName: "Test",
					LastName:  "User",
					Email:     "test_user1@domain.com",
					UserID:    "test_user2",
				},
			},
			wantErr: true,
			wantID:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{}
			err := u.FindOrCreateFromAuthUser(tt.args.orgID, tt.args.authUser)
			if err != nil && !tt.wantErr {
				t.Errorf("FindOrCreateFromAuthUser() error = %v, wantErr %v", err, tt.wantErr)
			} else if u.ID != tt.wantID {
				t.Errorf("ID on user is not what was wanted. Got %v, wanted %v", u.ID, tt.wantID)
			}
		})
	}
}

func (ms *ModelSuite) TestFindUserByAccessToken() {
	t := ms.T()
	resetTables(t)
	_, users, userOrgs := CreateUserFixtures(t)

	// Load access token test fixtures
	tokens := UserAccessTokens{
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken("abc123"),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken("xyz789"),
			ExpiresAt:          time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	if err := CreateUserAccessTokens(tokens); err != nil {
		t.Errorf("could not create access tokens ... %v", err)
		t.FailNow()
	}

	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    User
		wantErr bool
	}{
		{
			name:    "expired",
			args:    args{"abc123"},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{"xyz789"},
			want: users[0],
		},
		{
			name:    "invalid",
			args:    args{"000000"},
			wantErr: true,
		},
		{
			name:    "empty",
			args:    args{""},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := FindUserByAccessToken(test.args.token)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("FindUserByAccessToken() returned an error: %v", err)
				} else if got.Uuid != test.want.Uuid {
					t.Errorf("found %v, expected %v", got, test.want)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestValidateUser() {
	t := ms.T()
	tests := []struct {
		name     string
		user     User
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
				Uuid:      domain.GetUuid(),
			},
			wantErr: false,
		},
		{
			name: "missing email",
			user: User{
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
				Uuid:      domain.GetUuid(),
			},
			wantErr:  true,
			errField: "email",
		},
		{
			name: "missing first_name",
			user: User{
				Email:    "user@example.com",
				LastName: "User",
				Nickname: "A User",
				Uuid:     domain.GetUuid(),
			},
			wantErr:  true,
			errField: "first_name",
		},
		{
			name: "missing last_name",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				Nickname:  "A User",
				Uuid:      domain.GetUuid(),
			},
			wantErr:  true,
			errField: "last_name",
		},
		{
			name: "missing nickname",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Uuid:      domain.GetUuid(),
			},
			wantErr:  true,
			errField: "nickname",
		},
		{
			name: "missing uuid",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
			},
			wantErr:  true,
			errField: "uuid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.user.Validate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

// Ensure multiple access tokens for same organization are allowed (to support multiple tabs/browsers)
func (ms *ModelSuite) TestCreateAccessToken() {
	t := ms.T()
	resetTables(t)
	orgs, users, _ := CreateUserFixtures(t)

	type args struct {
		user     User
		clientID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "abc123",
			args: args{
				user:     users[0],
				clientID: "abc123",
			},
			wantErr: false,
		},
		{
			name: "123abc",
			args: args{
				user:     users[0],
				clientID: "123abc",
			},
			wantErr: false,
		},
		{
			name: "empty client ID",
			args: args{
				user:     users[0],
				clientID: "",
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedExpiry := createAccessTokenExpiry().Unix()
			token, expiry, err := test.args.user.CreateAccessToken(orgs[0], test.args.clientID)
			if test.wantErr {
				if err == nil {
					t.Errorf("expected error, but did not get one")
				}
			} else {
				if err != nil && !test.wantErr {
					t.Errorf("CreateAccessToken() returned error: %v", err)
				}
				hash := HashClientIdAccessToken(test.args.clientID + token)

				var dbToken UserAccessToken
				if err := DB.Where(fmt.Sprintf("access_token='%v'", hash)).First(&dbToken); err != nil {
					t.Errorf("Can't find new token (%v)", err)
				}

				if expiry-expectedExpiry > 1 {
					t.Errorf("Unexpected token expiry: %v, expected %v", expiry, expectedExpiry)
				}

				if dbToken.ExpiresAt.Unix()-expectedExpiry > 1 {
					t.Errorf("Unexpected token expiry: %v, expected %v", dbToken.ExpiresAt.Unix(), expectedExpiry)
				}
			}
		})
	}

	uat := &UserAccessToken{}
	count, _ := DB.Where("user_id = ?", users[0].ID).Count(uat)
	if count != 2 {
		t.Errorf("did not find correct number of user access tokens, want 2, got %v", count)
	}
}

func (ms *ModelSuite) TestGetOrgIDs() {
	t := ms.T()
	resetTables(t)
	_, users, _ := CreateUserFixtures(t)

	tests := []struct {
		name string
		user User
		want []int
	}{
		{
			name: "basic",
			user: users[0],
			want: []int{1, 2},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.user.GetOrgIDs()

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("GetOrgIDs() = \"%v\", want \"%v\"", got, test.want)
			}
		})
	}
}

func CreateUserFixtures(t *testing.T) (Organizations, Users, UserOrganizations) {
	resetTables(t)

	// Load Organization test fixtures
	orgs := Organizations{
		{
			Name:       "ACME",
			Uuid:       domain.GetUuid(),
			AuthType:   "saml2",
			AuthConfig: "[]",
		},
		{
			Name:       "Starfleet Academy",
			Uuid:       domain.GetUuid(),
			AuthType:   "saml2",
			AuthConfig: "[]",
		},
	}
	for i := range orgs {
		if err := DB.Create(&orgs[i]); err != nil {
			t.Errorf("error creating org %+v ...\n %v \n", orgs[i], err)
			t.FailNow()
		}
	}

	// Load User test fixtures
	users := Users{
		{
			Email:     "user1@example.com",
			FirstName: "Existing",
			LastName:  "User",
			Nickname:  "Existing User",
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "user2@example.com",
			FirstName: "Another",
			LastName:  "User",
			Nickname:  "Another User",
			Uuid:      domain.GetUuid(),
		},
	}
	for i := range users {
		if err := DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user %v ... %v", i, err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := UserOrganizations{
		{
			OrganizationID: orgs[0].ID,
			UserID:         users[0].ID,
			AuthID:         "existing_user",
			AuthEmail:      "user@example.com",
		},
		{
			OrganizationID: orgs[1].ID,
			UserID:         users[0].ID,
			AuthID:         "existing_user",
			AuthEmail:      "user@example.com",
		},
	}
	for i := range userOrgs {
		if err := DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	return orgs, users, userOrgs
}

func (ms *ModelSuite) TestGetOrganizations() {
	t := ms.T()
	resetTables(t)
	orgs, users, _ := CreateUserFixtures(t)

	tests := []struct {
		name string
		user User
		want []string
	}{
		{
			name: "basic",
			user: users[0],
			want: []string{orgs[0].Name, orgs[1].Name},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.user.GetOrganizations()
			if err != nil {
				t.Errorf("GetOrganizations() returned error: %s", err)
			}

			orgNames := make([]string, len(got))
			for i, o := range got {
				orgNames[i] = o.Name
			}
			if !reflect.DeepEqual(orgNames, test.want) {
				t.Errorf("GetOrgIDs() = \"%v\", want \"%v\"", got, test.want)
			}
		})
	}
}

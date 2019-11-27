package models

import (
	"strconv"
	"testing"

	"github.com/gobuffalo/validate"
	"github.com/silinternational/wecarry-api/domain"
)

type UserPreferenceFixtures struct {
	Users
	UserPreferences
}

func (ms *ModelSuite) TestUserPreference_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		pref     UserPreference
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			pref: UserPreference{
				Uuid:   domain.GetUuid(),
				UserID: 1,
				Key:    "key",
				Value:  "value",
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			pref: UserPreference{
				UserID: 1,
				Key:    "key",
				Value:  "value",
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "missing user_id",
			pref: UserPreference{
				Uuid:  domain.GetUuid(),
				Key:   "key",
				Value: "value",
			},
			wantErr:  true,
			errField: "user_id",
		},
		{
			name: "missing key",
			pref: UserPreference{
				Uuid:   domain.GetUuid(),
				UserID: 1,
				Value:  "value",
			},
			wantErr:  true,
			errField: "key",
		},
		{
			name: "missing value",
			pref: UserPreference{
				Uuid:   domain.GetUuid(),
				UserID: 1,
				Key:    "key",
			},
			wantErr:  true,
			errField: "value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.pref.Validate(DB)
			if test.wantErr {
				ms.True(vErr.Count() != 0, "Expected an error, but did not get one")
				ms.True(len(vErr.Get(test.errField)) > 0,
					"Expected an error on field %v, but got none (errors: %v)",
					test.errField, vErr.Errors)
				return
			}
			ms.False(vErr.HasAny(), "Unexpected error: %v", vErr)
		})
	}
}

func createFixturesForUserPreferenceFindByUUID(ms *ModelSuite) UserPreferenceFixtures {
	unique := domain.GetUuid().String()
	user := User{Uuid: domain.GetUuid(), Email: unique + "_user@example.com", Nickname: unique + "_User"}
	createFixture(ms, &user)

	userPreferences := make(UserPreferences, 2)
	for i := range userPreferences {
		userPreferences[i] = UserPreference{
			Uuid:   domain.GetUuid(),
			UserID: user.ID,
			Key:    "k",
			Value:  "v",
		}
		createFixture(ms, &userPreferences[i])
	}

	return UserPreferenceFixtures{
		Users:           Users{user},
		UserPreferences: userPreferences,
	}
}

func (ms *ModelSuite) TestUserPreference_FindByUUID() {
	t := ms.T()
	f := createFixturesForUserPreferenceFindByUUID(ms)
	tests := []struct {
		name    string
		uuid    string
		wantErr string
	}{
		{name: "good", uuid: f.UserPreferences[0].Uuid.String()},
		{name: "bad", wantErr: "user preference uuid must not be blank"},
		{name: "not found", uuid: domain.GetUuid().String(), wantErr: "sql: no rows in result set"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var u UserPreference
			err := u.FindByUUID(test.uuid)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}
			ms.NoError(err)
			ms.Equal(test.uuid, u.Uuid.String())
		})
	}
}

func createFixturesForUserPreferenceSave(ms *ModelSuite) UserPreferenceFixtures {
	unique := domain.GetUuid().String()
	user := User{Uuid: domain.GetUuid(), Email: unique + "_user@example.com", Nickname: unique + "_User"}
	createFixture(ms, &user)

	userPreferences := make(UserPreferences, 2)
	for i := range userPreferences {
		userPreferences[i] = UserPreference{
			UserID: user.ID,
			Key:    "key" + strconv.Itoa(i),
			Value:  "v",
		}
	}
	createFixture(ms, &userPreferences[0])

	return UserPreferenceFixtures{
		Users:           Users{user},
		UserPreferences: userPreferences,
	}
}

func (ms *ModelSuite) TestUserPreference_Save() {
	t := ms.T()
	f := createFixturesForUserPreferenceSave(ms)
	tests := []struct {
		name    string
		pref    UserPreference
		wantErr string
	}{
		{name: "update", pref: f.UserPreferences[0]},
		{name: "create", pref: f.UserPreferences[1]},
		{name: "bad", wantErr: "can not be blank"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.pref.Save()
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}
			ms.NoError(err)

			var u UserPreference
			ms.NoError(u.FindByUUID(test.pref.Uuid.String()))
			ms.Equal(test.pref.UserID, u.UserID)
			ms.Equal(test.pref.Key, u.Key)
			ms.Equal(test.pref.Value, u.Value)
		})
	}
}

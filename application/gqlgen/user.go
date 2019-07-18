package gqlgen

import (
	"strconv"

	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

// ConvertDBUserToGqlUser does what its name says, but also ...
func ConvertDBUserToGqlUser(dbUser models.User) (User, error) {
	dbID := strconv.Itoa(dbUser.ID)

	r := GetStringFromNullsString(dbUser.AdminRole)
	gqlRole := Role(*r)

	newGqlUser := User{
		ID:        dbID,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		Nickname:  dbUser.Nickname,
		AdminRole: &gqlRole,
		CreatedAt: domain.ConvertTimeToStringPtr(dbUser.CreatedAt),
		UpdatedAt: domain.ConvertTimeToStringPtr(dbUser.UpdatedAt),
	}

	return newGqlUser, nil
}

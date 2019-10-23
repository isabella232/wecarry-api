package models

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

type PostFixtures struct {
	Users
	Posts
}

func CreateFixturesValidateUpdate(ms *ModelSuite, t *testing.T) []Post {

	// Create org
	org := &Organization{
		ID:         1,
		Name:       "TestOrg",
		Url:        nulls.String{},
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	createFixture(t, org)

	// Create User
	user := User{
		Email:     "user1@example.com",
		FirstName: "Existing",
		LastName:  "User",
		Nickname:  "Existing User ",
		Uuid:      domain.GetUuid(),
	}

	if err := ms.DB.Create(&user); err != nil {
		t.Errorf("could not create test user %v ... %v", user.Email, err)
		t.FailNow()
	}

	// Load Post test fixtures
	posts := []Post{
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Open Request 0",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Committed Request 1",
			Size:           PostSizeMedium,
			Status:         PostStatusCommitted,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Accepted Request 2",
			Size:           PostSizeMedium,
			Status:         PostStatusAccepted,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Received Request 3",
			Size:           PostSizeMedium,
			Status:         PostStatusReceived,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Completed Request 4",
			Size:           PostSizeMedium,
			Status:         PostStatusCompleted,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Removed Request 5",
			Size:           PostSizeMedium,
			Status:         PostStatusRemoved,
			Uuid:           domain.GetUuid(),
		},
	}

	if err := CreatePosts(posts); err != nil {
		t.Errorf("could not create test post ... %v", err)
		t.FailNow()
	}
	return posts
}

func CreatePostFixtures(ms *ModelSuite, t *testing.T, users Users) []Post {
	if err := DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	// Load Post test fixtures
	posts := []Post{
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeRequest,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "A Request",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
		},
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeOffer,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "An Offer",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[1].ID),
		},
	}
	for i := range posts {
		if err := ms.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
		if err := DB.Load(&posts[i], "CreatedBy", "Provider", "Receiver", "Organization"); err != nil {
			t.Errorf("Error loading post associations: %s", err)
			t.FailNow()
		}
	}
	return posts
}

func CreateFixtures_Posts_FindByUser(ms *ModelSuite) PostFixtures {
	t := ms.T()

	orgs := Organizations{
		{Uuid: domain.GetUuid(), AuthConfig: "{}"},
		{Uuid: domain.GetUuid(), AuthConfig: "{}"},
	}
	for i := range orgs {
		createFixture(t, &orgs[i])
	}

	unique := domain.GetUuid().String()
	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "User0", Uuid: domain.GetUuid()},
		{Email: unique + "_user1@example.com", Nickname: unique + "User1", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(t, &users[i])
	}

	userOrgs := UserOrganizations{
		{OrganizationID: orgs[0].ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
		{OrganizationID: orgs[1].ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
		{OrganizationID: orgs[0].ID, UserID: users[1].ID, AuthID: users[1].Email, AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(t, &(userOrgs[i]))
	}

	posts := Posts{
		{Uuid: domain.GetUuid(), CreatedByID: users[0].ID, OrganizationID: orgs[0].ID},
		{Uuid: domain.GetUuid(), CreatedByID: users[0].ID, OrganizationID: orgs[1].ID},
		{Uuid: domain.GetUuid(), CreatedByID: users[1].ID, OrganizationID: orgs[0].ID},
	}
	for i := range posts {
		createFixture(t, &posts[i])
	}

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}
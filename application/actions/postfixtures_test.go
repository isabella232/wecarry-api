package actions

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type UpdatePostFixtures struct {
	models.Posts
	models.Users
	models.Files
	models.Locations
}

func createFixturesForUpdatePost(as *ActionSuite) UpdatePostFixtures {
	t := as.T()

	org := models.Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(as, &org)

	users := models.Users{
		{Email: t.Name() + "_user1@example.com", Nickname: t.Name() + " User1 ", Uuid: domain.GetUuid()},
		{Email: t.Name() + "_user2@example.com", Nickname: t.Name() + " User2 ", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(as, &users[i])
	}

	userOrgs := models.UserOrganizations{
		{OrganizationID: org.ID, UserID: users[0].ID, AuthID: t.Name() + "_auth_user1", AuthEmail: users[0].Email},
		{OrganizationID: org.ID, UserID: users[1].ID, AuthID: t.Name() + "_auth_user2", AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(as, &userOrgs[i])
	}

	accessTokenFixtures := []models.UserAccessToken{
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        models.HashClientIdAccessToken(users[0].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[1].ID,
			UserOrganizationID: userOrgs[1].ID,
			AccessToken:        models.HashClientIdAccessToken(users[1].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
	}
	for i := range accessTokenFixtures {
		createFixture(as, &accessTokenFixtures[i])
	}

	locations := []models.Location{
		{
			Description: "Miami, FL, USA",
			Country:     "US",
			Latitude:    nulls.NewFloat64(25.7617),
			Longitude:   nulls.NewFloat64(-80.1918),
		},
	}
	for i := range locations {
		createFixture(as, &locations[i])
	}

	posts := models.Posts{
		{
			CreatedByID:    users[0].ID,
			Type:           models.PostTypeRequest,
			OrganizationID: org.ID,
			Title:          "An Offer",
			Size:           models.PostSizeLarge,
			Status:         models.PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[1].ID),
			DestinationID:  locations[0].ID, // test update of existing location
			// leave OriginID nil to test adding a location
		},
	}

	for i := range posts {
		createFixture(as, &posts[i])
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	// create file fixtures
	fileData := []struct {
		name    string
		content []byte
	}{
		{
			name:    "photo.gif",
			content: []byte("GIF89a"),
		},
		{
			name:    "new_photo.webp",
			content: []byte("RIFFxxxxWEBPVP"),
		},
	}
	fileFixtures := make([]models.File, len(fileData))
	for i, fileDatum := range fileData {
		var f models.File
		if err := f.Store(fileDatum.name, fileDatum.content); err != nil {
			t.Errorf("failed to create file fixture, %s", err)
			t.FailNow()
		}
		fileFixtures[i] = f
	}

	// attach photo
	if _, err := posts[0].AttachPhoto(fileFixtures[0].UUID.String()); err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
		t.FailNow()
	}

	return UpdatePostFixtures{
		Posts: posts,
		Users: users,
		Files: fileFixtures,
	}
}

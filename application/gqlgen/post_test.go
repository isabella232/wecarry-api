package gqlgen

import (
	"strconv"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type PostQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Files
	models.Threads
	models.Locations
}

type PostResponse struct {
	Post struct {
		ID          string          `json:"id"`
		Type        models.PostType `json:"type"`
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Destination struct {
			Description string  `json:"description"`
			Country     string  `json:"country"`
			Lat         float64 `json:"latitude"`
			Long        float64 `json:"longitude"`
		} `json:"destination"`
		Origin struct {
			Description string  `json:"description"`
			Country     string  `json:"country"`
			Lat         float64 `json:"latitude"`
			Long        float64 `json:"longitude"`
		} `json:"origin"`
		Size         models.PostSize   `json:"size"`
		NeededAfter  string            `json:"neededAfter"`
		NeededBefore string            `json:"neededBefore"`
		Category     string            `json:"category"`
		Status       models.PostStatus `json:"status"`
		CreatedAt    string            `json:"createdAt"`
		UpdatedAt    string            `json:"updatedAt"`
		Cost         string            `json:"cost"`
		IsEditable   bool              `json:"isEditable"`
		Url          string            `json:"url"`
		CreatedBy    struct {
			ID string `json:"id"`
		} `json:"createdBy"`
		Receiver struct {
			ID string `json:"id"`
		} `json:"receiver"`
		Provider struct {
			ID string `json:"id"`
		} `json:"provider"`
		Organization struct {
			ID string `json:"id"`
		} `json:"organization"`
		Photo struct {
			ID string `json:"id"`
		} `json:"photo"`
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	} `json:"post"`
}

func createFixtures_PostQuery(gs *GqlgenSuite) PostQueryFixtures {
	t := gs.T()

	org := models.Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(gs, &org)

	users := models.Users{
		{Email: t.Name() + "_user1@example.com", Nickname: t.Name() + " User1 ", Uuid: domain.GetUuid()},
		{Email: t.Name() + "_user2@example.com", Nickname: t.Name() + " User2 ", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(gs, &users[i])
	}

	userOrgs := models.UserOrganizations{
		{OrganizationID: org.ID, UserID: users[0].ID, AuthID: t.Name() + "_auth_user1", AuthEmail: users[0].Email},
		{OrganizationID: org.ID, UserID: users[1].ID, AuthID: t.Name() + "_auth_user2", AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(gs, &userOrgs[i])
	}

	locations := []models.Location{
		{
			Description: "Miami, FL, USA",
			Country:     "US",
			Latitude:    nulls.NewFloat64(25.7617),
			Longitude:   nulls.NewFloat64(-80.1918),
		},
		{
			Description: "Toronto, Canada",
			Country:     "CA",
			Latitude:    nulls.NewFloat64(43.6532),
			Longitude:   nulls.NewFloat64(-79.3832),
		},
		{},
	}
	for i := range locations {
		createFixture(gs, &locations[i])
	}

	posts := models.Posts{
		{
			Uuid:           domain.GetUuid(),
			CreatedByID:    users[0].ID,
			ReceiverID:     nulls.NewInt(users[0].ID),
			ProviderID:     nulls.NewInt(users[1].ID),
			OrganizationID: org.ID,
			Type:           models.PostTypeRequest,
			Status:         models.PostStatusCommitted,
			Title:          "A Request",
			DestinationID:  locations[0].ID,
			OriginID:       nulls.NewInt(locations[1].ID),
			Size:           models.PostSizeSmall,
			NeededAfter:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			NeededBefore:   time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			Category:       "OTHER",
			Description:    nulls.NewString("This is a description"),
			URL:            nulls.NewString("https://www.example.com/items/101"),
			Cost:           nulls.NewFloat64(1.0),
		},
		{
			Uuid:           domain.GetUuid(),
			CreatedByID:    users[0].ID,
			ProviderID:     nulls.NewInt(users[0].ID),
			OrganizationID: org.ID,
			DestinationID:  locations[2].ID,
		},
	}
	for i := range posts {
		createFixture(gs, &posts[i])
	}

	threads := []models.Thread{
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(gs, &threads[i])
	}

	threadParticipants := []models.ThreadParticipant{
		{ThreadID: threads[0].ID, UserID: posts[0].CreatedByID},
	}
	for i := range threadParticipants {
		createFixture(gs, &threadParticipants[i])
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	fileData := []struct {
		name    string
		content []byte
	}{
		{"photo.gif", []byte("GIF89a")},
		{"dummy.pdf", []byte("%PDF-")},
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

	if _, err := posts[0].AttachPhoto(fileFixtures[0].UUID.String()); err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
		t.FailNow()
	}

	if _, err := posts[0].AttachFile(fileFixtures[1].UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
		t.FailNow()
	}

	return PostQueryFixtures{
		Organization: org,
		Users:        users,
		Posts:        posts,
		Files:        fileFixtures,
		Threads:      threads,
		Locations:    locations,
	}
}

func (gs *GqlgenSuite) Test_PostQuery() {
	f := createFixtures_PostQuery(gs)
	c := getGqlClient()

	query := `{ post(id: "` + f.Posts[0].Uuid.String() + `") 
		{ 
			id 
		    type
			title
			description
			destination {description country latitude longitude}
			origin {description country latitude longitude}
			size
			neededAfter
			neededBefore
			category
			status
			createdAt
			updatedAt
			cost
			isEditable
			url
			createdBy { id }
			receiver { id }
			provider { id }
			organization { id }
			photo { id }
			files { id } 
		}}`

	var resp PostResponse

	TestUser = f.Users[1]
	err := c.Post(query, &resp)
	gs.NoError(err)

	gs.Equal(f.Posts[0].Uuid.String(), resp.Post.ID)
	gs.Equal(f.Posts[0].Type, resp.Post.Type)
	gs.Equal(f.Posts[0].Title, resp.Post.Title)
	gs.Equal(f.Posts[0].Description.String, resp.Post.Description)

	gs.Equal(f.Locations[0].Description, resp.Post.Destination.Description)
	gs.Equal(f.Locations[0].Country, resp.Post.Destination.Country)
	gs.Equal(f.Locations[0].Latitude.Float64, resp.Post.Destination.Lat)
	gs.Equal(f.Locations[0].Longitude.Float64, resp.Post.Destination.Long)

	gs.Equal(f.Locations[1].Description, resp.Post.Origin.Description)
	gs.Equal(f.Locations[1].Country, resp.Post.Origin.Country)
	gs.Equal(f.Locations[1].Latitude.Float64, resp.Post.Origin.Lat)
	gs.Equal(f.Locations[1].Longitude.Float64, resp.Post.Origin.Long)

	gs.Equal(f.Posts[0].Size, resp.Post.Size)
	gs.Equal(f.Posts[0].NeededAfter.Format(time.RFC3339), resp.Post.NeededAfter)
	gs.Equal(f.Posts[0].NeededBefore.Format(time.RFC3339), resp.Post.NeededBefore)
	gs.Equal(f.Posts[0].Category, resp.Post.Category)
	gs.Equal(f.Posts[0].Status, resp.Post.Status)
	gs.Equal(f.Posts[0].CreatedAt.Format(time.RFC3339), resp.Post.CreatedAt)
	gs.Equal(f.Posts[0].UpdatedAt.Format(time.RFC3339), resp.Post.UpdatedAt)
	cost, err := strconv.ParseFloat(resp.Post.Cost, 64)
	gs.NoError(err, "couldn't parse cost field as a float")
	gs.Equal(f.Posts[0].URL.String, resp.Post.Url)
	gs.Equal(f.Posts[0].Cost.Float64, cost)
	gs.Equal(false, resp.Post.IsEditable)
	gs.Equal(f.Users[0].Uuid.String(), resp.Post.CreatedBy.ID)
	gs.Equal(f.Users[0].Uuid.String(), resp.Post.Receiver.ID)
	gs.Equal(f.Users[1].Uuid.String(), resp.Post.Provider.ID)
	gs.Equal(f.Organization.Uuid.String(), resp.Post.Organization.ID)
	gs.Equal(f.Files[0].UUID.String(), resp.Post.Photo.ID)
	gs.Equal(1, len(resp.Post.Files))
	gs.Equal(f.Files[1].UUID.String(), resp.Post.Files[0].ID)
}

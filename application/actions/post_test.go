package actions

import (
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type PostQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Threads
}

type PostsResponse struct {
	Posts []Post `json:"posts"`
}

type PostResponse struct {
	Post Post `json:"post"`
}

type PublicProfile struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarURL"`
}

type PotentialProvider struct {
	User           PublicProfile `json:"user"`
	DeliveryAfter  string        `json:"deliveryAfter"`
	DeliveryBefore string        `json:"deliveryBefore"`
}

type Post struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	NeededBefore *string `json:"neededBefore"`
	CompletedOn  *string `json:"completedOn"`
	Destination  struct {
		Description string  `json:"description"`
		Country     string  `json:"country"`
		Lat         float64 `json:"latitude"`
		Long        float64 `json:"longitude"`
	} `json:"destination"`
	Origin *struct {
		Description string  `json:"description"`
		Country     string  `json:"country"`
		Lat         float64 `json:"latitude"`
		Long        float64 `json:"longitude"`
	} `json:"origin"`
	Size       models.PostSize   `json:"size"`
	Status     models.PostStatus `json:"status"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
	Kilograms  *float64          `json:"kilograms"`
	IsEditable bool              `json:"isEditable"`
	Url        *string           `json:"url"`
	Visibility string            `json:"visibility"`
	CreatedBy  struct {
		ID        string `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarURL"`
	} `json:"createdBy"`
	Provider struct {
		ID        string `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarURL"`
	} `json:"provider"`
	PotentialProviders []PotentialProvider `json:"potentialProviders"`
	Organization       struct {
		ID string `json:"id"`
	} `json:"organization"`
	Photo *struct {
		ID string `json:"id"`
	} `json:"photo"`
	PhotoID string `json:"photoID"`
	Files   []struct {
		ID string `json:"id"`
	} `json:"files"`
	Meeting *struct {
		ID string `json:"id"`
	} `json:"meeting"`
}

const allPostFields = `{
			id
			createdBy { id nickname avatarURL }
			provider { id nickname avatarURL }
            potentialProviders { user { id nickname avatarURL }}
			organization { id }
			title
			description
			destination {description country latitude longitude}
			neededBefore
			completedOn
			destination {description country latitude longitude}
			origin {description country latitude longitude}
			size
			status
            threads { id }
			createdAt
			updatedAt
			url
			kilograms
			photo { id }
			photoID
			files { id }
            meeting { id }
			isEditable
			visibility
		}`

func (as *ActionSuite) Test_PostQuery() {
	f := createFixturesForPostQuery(as)

	query := fmt.Sprintf(`{ post(id: "%s") %s }`, f.Posts[0].UUID, allPostFields)

	var resp PostResponse

	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)

	as.Equal(f.Posts[0].UUID.String(), resp.Post.ID)
	as.Equal(f.Posts[0].Title, resp.Post.Title)
	as.Equal(f.Posts[0].Description.String, *resp.Post.Description)

	wantDate := f.Posts[0].NeededBefore.Time.Format(domain.DateFormat)
	as.Equal(wantDate, *resp.Post.NeededBefore, "incorrect NeededBefore date")
	as.Nil(resp.Post.CompletedOn, "expected a nil completedOn date string")

	as.NoError(as.DB.Load(&f.Posts[0], "Destination", "Origin", "PhotoFile", "Files.File"))

	as.Equal(f.Posts[0].Destination.Description, resp.Post.Destination.Description)
	as.Equal(f.Posts[0].Destination.Country, resp.Post.Destination.Country)
	as.Equal(f.Posts[0].Destination.Latitude.Float64, resp.Post.Destination.Lat)
	as.Equal(f.Posts[0].Destination.Longitude.Float64, resp.Post.Destination.Long)

	as.Equal(f.Posts[0].Origin.Description, resp.Post.Origin.Description)
	as.Equal(f.Posts[0].Origin.Country, resp.Post.Origin.Country)
	as.Equal(f.Posts[0].Origin.Latitude.Float64, resp.Post.Origin.Lat)
	as.Equal(f.Posts[0].Origin.Longitude.Float64, resp.Post.Origin.Long)

	as.Equal(f.Posts[0].Size, resp.Post.Size)
	as.Equal(f.Posts[0].Status, resp.Post.Status)
	as.Equal(f.Posts[0].CreatedAt.Format(time.RFC3339), resp.Post.CreatedAt)
	as.Equal(f.Posts[0].UpdatedAt.Format(time.RFC3339), resp.Post.UpdatedAt)
	as.Equal(f.Posts[0].Kilograms.Float64, *resp.Post.Kilograms)
	as.Equal(f.Posts[0].URL.String, *resp.Post.Url)
	as.Equal(f.Posts[0].Visibility.String(), resp.Post.Visibility)
	as.Equal(false, resp.Post.IsEditable)
	as.Equal(f.Users[0].UUID.String(), resp.Post.CreatedBy.ID, "creator ID doesn't match")
	as.Equal(f.Users[0].Nickname, resp.Post.CreatedBy.Nickname, "creator nickname doesn't match")
	as.Equal(f.Users[0].AuthPhotoURL.String, resp.Post.CreatedBy.AvatarURL, "creator avatar URL doesn't match")
	as.Equal(f.Users[1].UUID.String(), resp.Post.Provider.ID, "provider ID doesn't match")
	as.Equal(f.Users[1].Nickname, resp.Post.Provider.Nickname, "provider nickname doesn't match")
	as.Equal(f.Users[1].AuthPhotoURL.String, resp.Post.Provider.AvatarURL, "provider avatar URL doesn't match")
	as.Equal(f.Organization.UUID.String(), resp.Post.Organization.ID)
	as.Equal(f.Posts[0].PhotoFile.UUID.String(), resp.Post.Photo.ID)
	as.Equal(f.Posts[0].PhotoFile.UUID.String(), resp.Post.PhotoID)
	as.Equal(1, len(resp.Post.Files))
	as.Equal(f.Posts[0].Files[0].File.UUID.String(), resp.Post.Files[0].ID)

	// Check an actual CompletedOn field
	query = fmt.Sprintf(`{ post(id: "%s") %s }`, f.Posts[2].UUID, allPostFields)
	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)

	wantDate = f.Posts[2].CompletedOn.Time.Format(domain.DateFormat)
	as.Equal(wantDate, *resp.Post.CompletedOn, "incorrect CompletedOn date")
}

func (as *ActionSuite) Test_PostsQuery() {
	f := createFixturesForPostQuery(as)

	type testCase struct {
		Name        string
		SearchText  string
		Destination string
		Origin      string
		TestUser    models.User
		ExpectError bool
		Test        func(t *testing.T)
	}

	const queryTemplate = `{ posts (searchText: %s, destination: %s, origin: %s)` + allPostFields + `}`

	var resp PostsResponse

	testCases := []testCase{
		{
			Name:        "default",
			SearchText:  "null",
			Destination: "null",
			Origin:      "null",
			TestUser:    f.Users[1],
			Test: func(t *testing.T) {
				as.Equal(2, len(resp.Posts))
				as.Equal(f.Posts[0].UUID.String(), resp.Posts[1].ID)
				as.Equal(f.Posts[1].UUID.String(), resp.Posts[0].ID)
			},
		},
		{
			Name:        "searchText filter",
			SearchText:  `"title 0"`,
			Destination: "null",
			Origin:      "null",
			TestUser:    f.Users[1],
			Test: func(t *testing.T) {
				as.Equal(1, len(resp.Posts))
				as.Equal(f.Posts[0].UUID.String(), resp.Posts[0].ID)
			},
		},
		{
			Name:        "destination filter",
			SearchText:  "null",
			Destination: `{description:"Australia",country:"AU"}`,
			Origin:      "null",
			TestUser:    f.Users[1],
			Test: func(t *testing.T) {
				as.Equal(1, len(resp.Posts))
				as.Equal(f.Posts[0].UUID.String(), resp.Posts[0].ID)
			},
		},
		{
			Name:        "origin filter",
			SearchText:  "null",
			Destination: "null",
			Origin:      `{description:"Australia",country:"AU"}`,
			TestUser:    f.Users[1],
			Test: func(t *testing.T) {
				as.Equal(1, len(resp.Posts))
				as.Equal(f.Posts[1].UUID.String(), resp.Posts[0].ID)
			},
		},
	}

	for _, tc := range testCases {
		query := fmt.Sprintf(queryTemplate, tc.SearchText, tc.Destination, tc.Origin)
		err := as.testGqlQuery(query, tc.TestUser.Nickname, &resp)

		if tc.ExpectError {
			as.Error(err)
		} else {
			as.NoError(err)
		}
		as.T().Run(tc.Name, tc.Test)
	}
}

func (as *ActionSuite) Test_UpdatePost() {
	t := as.T()

	f := createFixturesForUpdatePost(as)

	var postsResp PostResponse

	input := `id: "` + f.Posts[0].UUID.String() + `" photoID: "` + f.Files[0].UUID.String() + `"` +
		`   title: "title"
			description: "new description"
			destination: {description:"dest" country:"dc" latitude:1.1 longitude:2.2}
			origin: {description:"origin" country:"oc" latitude:3.3 longitude:4.4}
			size: TINY
			url: "example.com"
			kilograms: 22.22
			neededBefore: "2099-12-31"
			visibility: ALL
		`
	query := `mutation { post: updatePost(input: {` + input + `}) { id photo { id } title description
			neededBefore
			destination { description country latitude longitude}
			origin { description country latitude longitude}
			size url kilograms visibility isEditable}}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))

	if err := as.DB.Load(&(f.Posts[0]), "PhotoFile", "Files"); err != nil {
		t.Errorf("failed to load post fixture, %s", err)
		t.FailNow()
	}

	as.Equal(f.Posts[0].UUID.String(), postsResp.Post.ID)
	as.Equal(f.Files[0].UUID.String(), postsResp.Post.Photo.ID)
	as.Equal("title", postsResp.Post.Title)
	as.Equal("new description", *postsResp.Post.Description)
	as.Equal("2099-12-31", *postsResp.Post.NeededBefore)
	as.Equal("dest", postsResp.Post.Destination.Description)
	as.Equal("dc", postsResp.Post.Destination.Country)
	as.Equal(1.1, postsResp.Post.Destination.Lat)
	as.Equal(2.2, postsResp.Post.Destination.Long)
	as.Equal("origin", postsResp.Post.Origin.Description)
	as.Equal("oc", postsResp.Post.Origin.Country)
	as.Equal(3.3, postsResp.Post.Origin.Lat)
	as.Equal(4.4, postsResp.Post.Origin.Long)
	as.Equal(models.PostSizeTiny, postsResp.Post.Size)
	as.Equal("example.com", *postsResp.Post.Url)
	as.Equal(22.22, *postsResp.Post.Kilograms)
	as.Equal("ALL", postsResp.Post.Visibility)
	as.Equal(true, postsResp.Post.IsEditable)

	// Attempt to edit a locked post
	input = `id: "` + f.Posts[0].UUID.String() + `" description: "new description"`
	query = `mutation { post: updatePost(input: {` + input + `}) { id status}}`

	as.Error(as.testGqlQuery(query, f.Users[1].Nickname, &postsResp))

	newNeededBefore := "2099-12-25"
	// Modify post's NeededBefore
	input = `id: "` + f.Posts[0].UUID.String() + `"
		neededBefore: "` + newNeededBefore + `"`
	query = `mutation { post: updatePost(input: {` + input + `}) { id neededBefore }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))
	as.Equal(newNeededBefore, *postsResp.Post.NeededBefore, "incorrect NeededBefore")

	// Null out post's nullable fields
	input = `id: "` + f.Posts[0].UUID.String() + `"`
	query = `mutation { post: updatePost(input: {` + input + `}) { id description url neededBefore kilograms
		photo { id } origin { description } meeting { id }  }}`

	postsResp = PostResponse{}
	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))
	as.Nil(postsResp.Post.Description, "Description is not nil")
	as.Nil(postsResp.Post.Url, "URL is not nil")
	as.Nil(postsResp.Post.NeededBefore, "NeededBefore is not nil")
	as.Nil(postsResp.Post.Kilograms, "Kilograms is not nil")
	as.Nil(postsResp.Post.Photo, "Photo is not nil")
	as.Nil(postsResp.Post.Origin, "Origin is not nil")
	as.Nil(postsResp.Post.Meeting, "Meeting is not nil")
}

func (as *ActionSuite) Test_CreatePost() {
	f := createFixturesForCreatePost(as)

	var postsResp PostResponse

	neededBefore := "2030-12-25"

	input := `orgID: "` + f.Organization.UUID.String() + `"` +
		`photoID: "` + f.File.UUID.String() + `"` +
		`
			title: "title"
			description: "new description"
			destination: {description:"dest" country:"dc" latitude:1.1 longitude:2.2}
			origin: {description:"origin" country:"oc" latitude:3.3 longitude:4.4}
			size: TINY
			url: "example.com"
			visibility: ALL
			kilograms: 1.5
		`
	query := `mutation { post: createPost(input: {` + input + `}) { organization { id } photo { id } title
			neededBefore description destination { description country latitude longitude }
			origin { description country latitude longitude }
			size url kilograms visibility }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))

	as.Equal(f.Organization.UUID.String(), postsResp.Post.Organization.ID)
	as.Equal(f.File.UUID.String(), postsResp.Post.Photo.ID)
	as.Equal("title", postsResp.Post.Title)
	as.Equal("new description", *postsResp.Post.Description)
	as.Nil(postsResp.Post.NeededBefore)
	as.Equal(models.PostStatus(""), postsResp.Post.Status)
	as.Equal("dest", postsResp.Post.Destination.Description)
	as.Equal("dc", postsResp.Post.Destination.Country)
	as.Equal(1.1, postsResp.Post.Destination.Lat)
	as.Equal(2.2, postsResp.Post.Destination.Long)
	as.Equal("origin", postsResp.Post.Origin.Description)
	as.Equal("oc", postsResp.Post.Origin.Country)
	as.Equal(3.3, postsResp.Post.Origin.Lat)
	as.Equal(4.4, postsResp.Post.Origin.Long)
	as.Equal(models.PostSizeTiny, postsResp.Post.Size)
	as.Equal("example.com", *postsResp.Post.Url)
	as.Equal(1.5, *postsResp.Post.Kilograms)
	as.Equal("ALL", postsResp.Post.Visibility)

	// meeting-based request
	input = `orgID: "` + f.Organization.UUID.String() + `"` +
		`meetingID: "` + f.Meetings[0].UUID.String() + `"` +
		`
			title: "title"
			description: "new description"
			neededBefore: "` + neededBefore + `"
			destination: {description:"dest" country:"dc" latitude:1.1 longitude:2.2}
			size: TINY
			url: "example.com"
		`
	query = `mutation { post: createPost(input: {` + input + `}) {
		neededBefore destination { description country latitude longitude }
		meeting { id } }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))

	as.Equal(f.Meetings[0].UUID.String(), postsResp.Post.Meeting.ID)

	as.NoError(as.DB.Load(&f.Meetings[0]), "Location")
	as.Equal(f.Meetings[0].Location.Description, postsResp.Post.Destination.Description)

	as.NotNil(postsResp.Post.NeededBefore)
	as.Equal(neededBefore, *postsResp.Post.NeededBefore)

	as.Equal(f.Meetings[0].Location.Country, postsResp.Post.Destination.Country)
	as.Equal(f.Meetings[0].Location.Latitude.Float64, postsResp.Post.Destination.Lat)
	as.Equal(f.Meetings[0].Location.Longitude.Float64, postsResp.Post.Destination.Long)
}

func (as *ActionSuite) Test_UpdatePostStatus() {
	f := createFixturesForUpdatePostStatus(as)

	var postsResp PostResponse

	creator := f.Users[0]
	provider := f.Users[1]

	steps := []struct {
		status     models.PostStatus
		user       models.User
		providerID string
		wantErr    bool
	}{
		{status: models.PostStatusAccepted, user: provider, providerID: provider.UUID.String(), wantErr: true},
		{status: models.PostStatusAccepted, user: creator, providerID: provider.UUID.String(), wantErr: false},
		{status: models.PostStatusReceived, user: provider, wantErr: true},
		{status: models.PostStatusReceived, user: creator, wantErr: false},
		{status: models.PostStatusDelivered, user: provider, wantErr: false},
		{status: models.PostStatusCompleted, user: provider, wantErr: true},
		{status: models.PostStatusCompleted, user: creator, wantErr: false},
		{status: models.PostStatusRemoved, user: creator, wantErr: true},
	}

	for _, step := range steps {
		input := `id: "` + f.Posts[0].UUID.String() + `", status: ` + step.status.String()
		if step.providerID != "" {
			input += `, providerUserID: "` + step.providerID + `"`
		}
		query := `mutation { post: updatePostStatus(input: {` + input + `}) {id status completedOn}}`

		err := as.testGqlQuery(query, step.user.Nickname, &postsResp)
		if step.wantErr {
			as.Error(err, "user=%s, query=%s", step.user.Nickname, query)
			continue
		}

		as.NoError(err, "user=%s, query=%s", step.user.Nickname, query)
		as.Equal(step.status, postsResp.Post.Status)

		if step.status == models.PostStatusCompleted {
			as.NotNil(postsResp.Post.CompletedOn, "expected valid CompletedOn date.")
			as.Equal(time.Now().Format(domain.DateFormat), *postsResp.Post.CompletedOn,
				"incorrect CompletedOn date.")
		} else {
			as.Nil(postsResp.Post.CompletedOn, "expected nil CompletedOn field.")
		}
	}
}

func (as *ActionSuite) Test_UpdatePostStatus_DestroyPotentialProviders() {
	f := test.CreatePotentialProvidersFixtures(as.DB)
	users := f.Users

	var postsResp PostResponse

	creator := f.Users[0]
	provider := f.Users[1]

	post0 := f.Posts[0]
	post0.Status = models.PostStatusAccepted
	err := post0.Update()
	as.NoError(err, "unable to change Posts's status to prepare for test")

	steps := []struct {
		status    models.PostStatus
		user      models.User
		wantPPIDs []uuid.UUID
		wantErr   bool
	}{
		{status: models.PostStatusReceived, user: creator,
			wantPPIDs: []uuid.UUID{users[1].UUID, users[2].UUID, users[3].UUID}},
		{status: models.PostStatusDelivered, user: creator,
			wantPPIDs: []uuid.UUID{users[1].UUID, users[2].UUID, users[3].UUID}},
		{status: models.PostStatusCompleted, user: provider, wantErr: true},
		{status: models.PostStatusCompleted, user: creator, wantPPIDs: []uuid.UUID{}},
	}

	for _, step := range steps {
		input := `id: "` + f.Posts[0].UUID.String() + `", status: ` + step.status.String()
		query := `mutation { post: updatePostStatus(input: {` + input + `}) {id status potentialProviders {user {nickname}}}}`

		err := as.testGqlQuery(query, step.user.Nickname, &postsResp)
		if step.wantErr {
			as.Error(err, "user=%s, query=%s", step.user.Nickname, query)
		} else {
			as.NoError(err, "user=%s, query=%s", step.user.Nickname, query)
			as.Equal(step.status, postsResp.Post.Status)
		}
	}
}

func (as *ActionSuite) Test_MarkRequestAsDelivered() {
	f := createFixturesForMarkRequestAsDelivered(as)
	posts := f.Posts

	var postsResp PostResponse

	creator := f.Users[0]
	provider := f.Users[1]

	testCases := []struct {
		name                 string
		postID               string
		user                 models.User
		wantStatus           string
		wantPostHistoryCount int
		wantErr              bool
		wantErrContains      string
	}{
		{name: "ACCEPTED: delivered by Provider",
			postID: posts[0].UUID.String(), user: provider,
			wantStatus:           models.PostStatusDelivered.String(),
			wantPostHistoryCount: 3, wantErr: false},
		{name: "ACCEPTED: delivered by Creator",
			postID: posts[0].UUID.String(), user: creator, wantErr: true,
			wantErrContains: "not allowed to change the status",
		},
		{name: "COMPLETED: delivered by Provider",
			postID: posts[1].UUID.String(), user: provider, wantErr: true,
			wantErrContains: "not allowed to change the status",
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			query := fmt.Sprintf(`mutation { post: markRequestAsDelivered(postID: "%v") {id status completedOn}}`,
				tc.postID)

			err := as.testGqlQuery(query, tc.user.Nickname, &postsResp)
			if tc.wantErr {
				as.Error(err, "user=%d, query=%s", tc.user.ID, query)
				as.Contains(err.Error(), tc.wantErrContains, "incorrect error message")
				return
			}

			as.NoError(err, "user=%d, query=%s", tc.user.ID, query)
			as.Equal(tc.wantStatus, postsResp.Post.Status.String(), "incorrect status")

			if tc.wantPostHistoryCount < 1 {
				return
			}

			// Check for correct PostHistory
			var post models.Post
			as.NoError(post.FindByUUID(tc.postID))
			pHistories := models.PostHistories{}
			err = as.DB.Where("post_id = ?", post.ID).All(&pHistories)
			as.NoError(err)
			as.Equal(tc.wantPostHistoryCount, len(pHistories), "incorrect number of PostHistories")
			lastPH := pHistories[tc.wantPostHistoryCount-1]
			as.Equal(tc.wantStatus, lastPH.Status.String(), "incorrect status on last PostHistory")
		})
	}
}

func (as *ActionSuite) Test_MarkRequestAsReceived() {
	f := createFixturesForMarkRequestAsReceived(as)
	posts := f.Posts

	var postsResp PostResponse

	creator := f.Users[0]
	provider := f.Users[1]

	testCases := []struct {
		name                 string
		postID               string
		user                 models.User
		wantStatus           string
		wantPostHistoryCount int
		wantErr              bool
		wantErrContains      string
	}{
		{name: "ACCEPTED: received by Provider",
			postID: posts[0].UUID.String(), user: provider, wantErr: true,
			wantErrContains: "not allowed to change the status",
		},
		{name: "ACCEPTED: received by Creator",
			postID: posts[0].UUID.String(), user: creator, wantErr: false,
			wantStatus:           models.PostStatusCompleted.String(),
			wantPostHistoryCount: 3,
		},
		{name: "DELIVERED: received by Creator",
			postID: posts[1].UUID.String(), user: creator, wantErr: false,
			wantStatus:           models.PostStatusCompleted.String(),
			wantPostHistoryCount: 4,
		},
		{name: "COMPLETED: received by Creator",
			postID: posts[2].UUID.String(), user: creator, wantErr: true,
			wantErrContains: "not allowed to change the status",
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			query := fmt.Sprintf(`mutation { post: markRequestAsReceived(postID: "%v") {id status completedOn}}`,
				tc.postID)

			err := as.testGqlQuery(query, tc.user.Nickname, &postsResp)
			if tc.wantErr {
				as.Error(err, "user=%d, query=%s", tc.user.ID, query)
				as.Contains(err.Error(), tc.wantErrContains, "incorrect error message")
				return
			}

			as.NoError(err, "user=%d, query=%s", tc.user.ID, query)
			as.Equal(tc.wantStatus, postsResp.Post.Status.String(), "incorrect status")

			if tc.wantPostHistoryCount < 1 {
				return
			}

			// Check for correct PostHistory
			var post models.Post
			as.NoError(post.FindByUUID(tc.postID))
			pHistories := models.PostHistories{}
			err = as.DB.Where("post_id = ?", post.ID).All(&pHistories)
			as.NoError(err)
			as.Equal(tc.wantPostHistoryCount, len(pHistories), "incorrect number of PostHistories")
			lastPH := pHistories[tc.wantPostHistoryCount-1]
			as.Equal(tc.wantStatus, lastPH.Status.String(), "incorrect status on last PostHistory")
		})
	}
}

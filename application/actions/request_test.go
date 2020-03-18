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

type RequestQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Threads
}

type RequestsResponse struct {
	Requests []Request `json:"requests"`
}

type RequestResponse struct {
	Request Request `json:"request"`
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

type Request struct {
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

const allRequestFields = `{
			id
			createdBy { id nickname avatarURL }
			provider { id nickname avatarURL }
            potentialProviders { user {id nickname avatarURL} deliveryAfter deliveryBefore}
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

func (as *ActionSuite) verifyRequestResponse(request models.Post, resp Request) {
	as.Equal(request.UUID.String(), resp.ID)
	as.Equal(request.Title, resp.Title)
	as.Equal(request.Description.String, *resp.Description)

	wantDate := request.NeededBefore.Time.Format(domain.DateFormat)
	as.Equal(wantDate, *resp.NeededBefore, "incorrect NeededBefore date")
	as.Nil(resp.CompletedOn, "expected a nil completedOn date string")

	as.NoError(as.DB.Load(&request, "Destination", "Origin", "PhotoFile", "Files.File"))

	as.Equal(request.Destination.Description, resp.Destination.Description)
	as.Equal(request.Destination.Country, resp.Destination.Country)
	as.Equal(request.Destination.Latitude.Float64, resp.Destination.Lat)
	as.Equal(request.Destination.Longitude.Float64, resp.Destination.Long)

	as.Equal(request.Origin.Description, resp.Origin.Description)
	as.Equal(request.Origin.Country, resp.Origin.Country)
	as.Equal(request.Origin.Latitude.Float64, resp.Origin.Lat)
	as.Equal(request.Origin.Longitude.Float64, resp.Origin.Long)

	as.Equal(request.Size, resp.Size)
	as.Equal(request.Status, resp.Status)
	as.Equal(request.CreatedAt.Format(time.RFC3339), resp.CreatedAt)
	as.Equal(request.UpdatedAt.Format(time.RFC3339), resp.UpdatedAt)
	as.Equal(request.Kilograms.Float64, *resp.Kilograms)
	as.Equal(request.URL.String, *resp.Url)
	as.Equal(request.Visibility.String(), resp.Visibility)
	as.Equal(false, resp.IsEditable)

	creator, err := request.GetCreator()
	as.NoError(err)
	as.Equal(creator.UUID.String(), resp.CreatedBy.ID, "creator ID doesn't match")
	as.Equal(creator.Nickname, resp.CreatedBy.Nickname, "creator nickname doesn't match")
	as.Equal(creator.AuthPhotoURL.String, resp.CreatedBy.AvatarURL, "creator avatar URL doesn't match")

	provider, err := request.GetProvider()
	as.NoError(err)
	if provider != nil {
		as.Equal(provider.UUID.String(), resp.Provider.ID, "provider ID doesn't match")
		as.Equal(provider.Nickname, resp.Provider.Nickname, "provider nickname doesn't match")
		as.Equal(provider.AuthPhotoURL.String, resp.Provider.AvatarURL, "provider avatar URL doesn't match")
	}

	org, err := request.GetOrganization()
	as.NoError(err)
	as.Equal(org.UUID.String(), resp.Organization.ID)

	as.Equal(request.PhotoFile.UUID.String(), resp.Photo.ID)
	as.Equal(request.PhotoFile.UUID.String(), resp.PhotoID)
	as.Equal(len(request.Files), len(resp.Files))
	for i := range request.Files {
		as.Equal(request.Files[i].File.UUID.String(), resp.Files[i].ID)
	}

	if resp.CompletedOn != nil {
		wantDate = request.CompletedOn.Time.Format(domain.DateFormat)
		as.Equal(wantDate, *resp.CompletedOn, "incorrect CompletedOn date")
	}
}

func (as *ActionSuite) Test_RequestsQuery() {
	f := createFixturesForRequestQuery(as)

	type testCase struct {
		name        string
		searchText  string
		destination string
		origin      string
		testUser    models.User
		expectError bool
		verifyFunc  func()
	}

	const queryTemplate = `{ requests (searchText: %s, destination: %s, origin: %s)` + allRequestFields + `}`

	var resp RequestsResponse

	testCases := []testCase{
		{
			name:        "default",
			searchText:  "null",
			destination: "null",
			origin:      "null",
			testUser:    f.Users[1],
			verifyFunc: func() {
				as.Equal(2, len(resp.Requests))
				as.verifyRequestResponse(f.Posts[0], resp.Requests[1])
				as.verifyRequestResponse(f.Posts[1], resp.Requests[0])
			},
		},
		{
			name:        "searchText filter",
			searchText:  `"title 0"`,
			destination: "null",
			origin:      "null",
			testUser:    f.Users[1],
			verifyFunc: func() {
				as.Equal(1, len(resp.Requests))
				as.verifyRequestResponse(f.Posts[0], resp.Requests[0])
			},
		},
		{
			name:        "destination filter",
			searchText:  "null",
			destination: `{description:"Australia",country:"AU"}`,
			origin:      "null",
			testUser:    f.Users[1],
			verifyFunc: func() {
				as.Equal(1, len(resp.Requests))
				as.verifyRequestResponse(f.Posts[0], resp.Requests[0])
			},
		},
		{
			name:        "origin filter",
			searchText:  "null",
			destination: "null",
			origin:      `{description:"Australia",country:"AU"}`,
			testUser:    f.Users[1],
			verifyFunc: func() {
				as.Equal(1, len(resp.Requests))
				as.verifyRequestResponse(f.Posts[1], resp.Requests[0])
			},
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			query := fmt.Sprintf(queryTemplate, tc.searchText, tc.destination, tc.origin)
			resp = RequestsResponse{}
			err := as.testGqlQuery(query, tc.testUser.Nickname, &resp)

			if tc.expectError {
				as.Error(err)
				return
			}
			as.NoError(err)
			tc.verifyFunc()
		})
	}
}

func (as *ActionSuite) Test_UpdateRequest() {
	t := as.T()

	f := createFixturesForUpdateRequest(as)

	var requestsResp RequestResponse

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
	query := `mutation { request: updateRequest(input: {` + input + `}) { id photo { id } title description
			neededBefore
			destination { description country latitude longitude}
			origin { description country latitude longitude}
			size url kilograms visibility isEditable}}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &requestsResp))

	if err := as.DB.Load(&(f.Posts[0]), "PhotoFile", "Files"); err != nil {
		t.Errorf("failed to load request fixture, %s", err)
		t.FailNow()
	}

	as.Equal(f.Posts[0].UUID.String(), requestsResp.Request.ID)
	as.Equal(f.Files[0].UUID.String(), requestsResp.Request.Photo.ID)
	as.Equal("title", requestsResp.Request.Title)
	as.Equal("new description", *requestsResp.Request.Description)
	as.Equal("2099-12-31", *requestsResp.Request.NeededBefore)
	as.Equal("dest", requestsResp.Request.Destination.Description)
	as.Equal("dc", requestsResp.Request.Destination.Country)
	as.Equal(1.1, requestsResp.Request.Destination.Lat)
	as.Equal(2.2, requestsResp.Request.Destination.Long)
	as.Equal("origin", requestsResp.Request.Origin.Description)
	as.Equal("oc", requestsResp.Request.Origin.Country)
	as.Equal(3.3, requestsResp.Request.Origin.Lat)
	as.Equal(4.4, requestsResp.Request.Origin.Long)
	as.Equal(models.PostSizeTiny, requestsResp.Request.Size)
	as.Equal("example.com", *requestsResp.Request.Url)
	as.Equal(22.22, *requestsResp.Request.Kilograms)
	as.Equal("ALL", requestsResp.Request.Visibility)
	as.Equal(true, requestsResp.Request.IsEditable)

	// Attempt to edit a locked request
	input = `id: "` + f.Posts[0].UUID.String() + `" description: "new description"`
	query = `mutation { request: updateRequest(input: {` + input + `}) { id status}}`

	as.Error(as.testGqlQuery(query, f.Users[1].Nickname, &requestsResp))

	newNeededBefore := "2099-12-25"
	// Modify request's NeededBefore
	input = `id: "` + f.Posts[0].UUID.String() + `"
		neededBefore: "` + newNeededBefore + `"`
	query = `mutation { request: updateRequest(input: {` + input + `}) { id neededBefore }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &requestsResp))
	as.Equal(newNeededBefore, *requestsResp.Request.NeededBefore, "incorrect NeededBefore")

	// Null out request's nullable fields
	input = `id: "` + f.Posts[0].UUID.String() + `"`
	query = `mutation { request: updateRequest(input: {` + input + `}) { id description url neededBefore kilograms
		photo { id } origin { description } meeting { id }  }}`

	requestsResp = RequestResponse{}
	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &requestsResp))
	as.Nil(requestsResp.Request.Description, "Description is not nil")
	as.Nil(requestsResp.Request.Url, "URL is not nil")
	as.Nil(requestsResp.Request.NeededBefore, "NeededBefore is not nil")
	as.Nil(requestsResp.Request.Kilograms, "Kilograms is not nil")
	as.Nil(requestsResp.Request.Photo, "Photo is not nil")
	as.Nil(requestsResp.Request.Origin, "Origin is not nil")
	as.Nil(requestsResp.Request.Meeting, "Meeting is not nil")
}

func (as *ActionSuite) Test_CreateRequest() {
	f := createFixturesForCreateRequest(as)

	var requestsResp RequestResponse

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
	query := `mutation { request: createRequest(input: {` + input + `}) { organization { id } photo { id } title
			neededBefore description destination { description country latitude longitude }
			origin { description country latitude longitude }
			size url kilograms visibility }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &requestsResp))

	as.Equal(f.Organization.UUID.String(), requestsResp.Request.Organization.ID)
	as.Equal(f.File.UUID.String(), requestsResp.Request.Photo.ID)
	as.Equal("title", requestsResp.Request.Title)
	as.Equal("new description", *requestsResp.Request.Description)
	as.Nil(requestsResp.Request.NeededBefore)
	as.Equal(models.PostStatus(""), requestsResp.Request.Status)
	as.Equal("dest", requestsResp.Request.Destination.Description)
	as.Equal("dc", requestsResp.Request.Destination.Country)
	as.Equal(1.1, requestsResp.Request.Destination.Lat)
	as.Equal(2.2, requestsResp.Request.Destination.Long)
	as.Equal("origin", requestsResp.Request.Origin.Description)
	as.Equal("oc", requestsResp.Request.Origin.Country)
	as.Equal(3.3, requestsResp.Request.Origin.Lat)
	as.Equal(4.4, requestsResp.Request.Origin.Long)
	as.Equal(models.PostSizeTiny, requestsResp.Request.Size)
	as.Equal("example.com", *requestsResp.Request.Url)
	as.Equal(1.5, *requestsResp.Request.Kilograms)
	as.Equal("ALL", requestsResp.Request.Visibility)

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
	query = `mutation { request: createRequest(input: {` + input + `}) {
		neededBefore destination { description country latitude longitude }
		meeting { id } }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &requestsResp))

	as.Equal(f.Meetings[0].UUID.String(), requestsResp.Request.Meeting.ID)

	as.NoError(as.DB.Load(&f.Meetings[0]), "Location")
	as.Equal(f.Meetings[0].Location.Description, requestsResp.Request.Destination.Description)

	as.NotNil(requestsResp.Request.NeededBefore)
	as.Equal(neededBefore, *requestsResp.Request.NeededBefore)

	as.Equal(f.Meetings[0].Location.Country, requestsResp.Request.Destination.Country)
	as.Equal(f.Meetings[0].Location.Latitude.Float64, requestsResp.Request.Destination.Lat)
	as.Equal(f.Meetings[0].Location.Longitude.Float64, requestsResp.Request.Destination.Long)
}

func (as *ActionSuite) Test_UpdateRequestStatus() {
	f := createFixturesForUpdateRequestStatus(as)

	var requestsResp RequestResponse

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
		query := `mutation { request: updateRequestStatus(input: {` + input + `}) {id status completedOn}}`

		err := as.testGqlQuery(query, step.user.Nickname, &requestsResp)
		if step.wantErr {
			as.Error(err, "user=%s, query=%s", step.user.Nickname, query)
			continue
		}

		as.NoError(err, "user=%s, query=%s", step.user.Nickname, query)
		as.Equal(step.status, requestsResp.Request.Status)

		if step.status == models.PostStatusCompleted {
			as.NotNil(requestsResp.Request.CompletedOn, "expected valid CompletedOn date.")
			as.Equal(time.Now().Format(domain.DateFormat), *requestsResp.Request.CompletedOn,
				"incorrect CompletedOn date.")
		} else {
			as.Nil(requestsResp.Request.CompletedOn, "expected nil CompletedOn field.")
		}
	}
}

func (as *ActionSuite) Test_UpdateRequestStatus_DestroyPotentialProviders() {
	f := test.CreatePotentialProvidersFixtures(as.DB)
	users := f.Users

	var requestsResp RequestResponse

	creator := f.Users[0]
	provider := f.Users[1]

	request0 := f.Posts[0]
	request0.Status = models.PostStatusAccepted
	err := request0.Update()
	as.NoError(err, "unable to change Requests's status to prepare for test")

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
		query := `mutation { request: updateRequestStatus(input: {` + input + `}) {id status potentialProviders {user {nickname}}}}`

		err := as.testGqlQuery(query, step.user.Nickname, &requestsResp)
		if step.wantErr {
			as.Error(err, "user=%s, query=%s", step.user.Nickname, query)
		} else {
			as.NoError(err, "user=%s, query=%s", step.user.Nickname, query)
			as.Equal(step.status, requestsResp.Request.Status)
		}
	}
}

func (as *ActionSuite) Test_MarkRequestAsDelivered() {
	f := createFixturesForMarkRequestAsDelivered(as)
	requests := f.Posts

	var requestsResp RequestResponse

	creator := f.Users[0]
	provider := f.Users[1]

	testCases := []struct {
		name                    string
		requestID               string
		user                    models.User
		wantStatus              string
		wantRequestHistoryCount int
		wantErr                 bool
		wantErrContains         string
	}{
		{name: "ACCEPTED: delivered by Provider",
			requestID: requests[0].UUID.String(), user: provider,
			wantStatus:              models.PostStatusDelivered.String(),
			wantRequestHistoryCount: 3, wantErr: false},
		{name: "ACCEPTED: delivered by Creator",
			requestID: requests[0].UUID.String(), user: creator, wantErr: true,
			wantErrContains: "not allowed to change the status",
		},
		{name: "COMPLETED: delivered by Provider",
			requestID: requests[1].UUID.String(), user: provider, wantErr: true,
			wantErrContains: "not allowed to change the status",
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			query := fmt.Sprintf(`mutation { request: markRequestAsDelivered(requestID: "%v") {id status completedOn}}`,
				tc.requestID)

			err := as.testGqlQuery(query, tc.user.Nickname, &requestsResp)
			if tc.wantErr {
				as.Error(err, "user=%d, query=%s", tc.user.ID, query)
				as.Contains(err.Error(), tc.wantErrContains, "incorrect error message")
				return
			}

			as.NoError(err, "user=%d, query=%s", tc.user.ID, query)
			as.Equal(tc.wantStatus, requestsResp.Request.Status.String(), "incorrect status")

			if tc.wantRequestHistoryCount < 1 {
				return
			}

			// Check for correct RequestHistory
			var request models.Post
			as.NoError(request.FindByUUID(tc.requestID))
			pHistories := models.PostHistories{}
			err = as.DB.Where("post_id = ?", request.ID).All(&pHistories)
			as.NoError(err)
			as.Equal(tc.wantRequestHistoryCount, len(pHistories), "incorrect number of RequestHistories")
			lastPH := pHistories[tc.wantRequestHistoryCount-1]
			as.Equal(tc.wantStatus, lastPH.Status.String(), "incorrect status on last RequestHistory")
		})
	}
}

func (as *ActionSuite) Test_MarkRequestAsReceived() {
	f := createFixturesForMarkRequestAsReceived(as)
	requests := f.Posts

	var requestsResp RequestResponse

	creator := f.Users[0]
	provider := f.Users[1]

	testCases := []struct {
		name                    string
		requestID               string
		user                    models.User
		wantStatus              string
		wantRequestHistoryCount int
		wantErr                 bool
		wantErrContains         string
	}{
		{name: "ACCEPTED: received by Provider",
			requestID: requests[0].UUID.String(), user: provider, wantErr: true,
			wantErrContains: "not allowed to change the status",
		},
		{name: "ACCEPTED: received by Creator",
			requestID: requests[0].UUID.String(), user: creator, wantErr: false,
			wantStatus:              models.PostStatusCompleted.String(),
			wantRequestHistoryCount: 3,
		},
		{name: "DELIVERED: received by Creator",
			requestID: requests[1].UUID.String(), user: creator, wantErr: false,
			wantStatus:              models.PostStatusCompleted.String(),
			wantRequestHistoryCount: 4,
		},
		{name: "COMPLETED: received by Creator",
			requestID: requests[2].UUID.String(), user: creator, wantErr: true,
			wantErrContains: "not allowed to change the status",
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			query := fmt.Sprintf(`mutation { request: markRequestAsReceived(requestID: "%v") {id status completedOn}}`,
				tc.requestID)

			err := as.testGqlQuery(query, tc.user.Nickname, &requestsResp)
			if tc.wantErr {
				as.Error(err, "user=%d, query=%s", tc.user.ID, query)
				as.Contains(err.Error(), tc.wantErrContains, "incorrect error message")
				return
			}

			as.NoError(err, "user=%d, query=%s", tc.user.ID, query)
			as.Equal(tc.wantStatus, requestsResp.Request.Status.String(), "incorrect status")

			if tc.wantRequestHistoryCount < 1 {
				return
			}

			// Check for correct RequestHistory
			var request models.Post
			as.NoError(request.FindByUUID(tc.requestID))
			pHistories := models.PostHistories{}
			err = as.DB.Where("post_id = ?", request.ID).All(&pHistories)
			as.NoError(err)
			as.Equal(tc.wantRequestHistoryCount, len(pHistories), "incorrect number of RequestHistories")
			lastPH := pHistories[tc.wantRequestHistoryCount-1]
			as.Equal(tc.wantStatus, lastPH.Status.String(), "incorrect status on last RequestHistory")
		})
	}
}

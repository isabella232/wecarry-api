package models

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"

	"github.com/silinternational/wecarry-api/domain"
)

type UserFixtures struct {
	Organization
	Users
	UserOrganizations
	UserAccessTokens
	Locations
}

// mustCreate saves a record to the database. Panics if any error occurs.
func mustCreate(tx *pop.Connection, f interface{}) {
	value := reflect.ValueOf(f)

	if value.Type().Kind() != reflect.Ptr {
		panic("mustCreate requires a pointer")
	}

	uuidField := value.Elem().FieldByName("UUID")
	if uuidField.IsValid() {
		uuidField.Set(reflect.ValueOf(domain.GetUUID()))
	}

	err := tx.Create(f)
	if err != nil {
		panic(fmt.Sprintf("error creating %T fixture, %s", f, err))
	}
}

// createOrganizationFixtures generates any number of organization records for testing.
//  Their names will be called "Org1", "Org2", ...
func createOrganizationFixtures(tx *pop.Connection, n int) Organizations {
	files := make([]File, n)
	organizations := make(Organizations, n)
	for i := range organizations {
		if err := files[i].Store("logo.gif", []byte("GIF89a")); err != nil {
			panic("error storing org logo, " + err.Error())
		}
		organizations[i].Name = fmt.Sprintf("Org%v", i+1)
		organizations[i].AuthType = AuthTypeSaml
		organizations[i].AuthConfig = "{}"
		if _, err := organizations[i].AttachLogo(files[i].UUID.String()); err != nil {
			panic("error attaching logo to org fixture, " + err.Error())
		}

		mustCreate(tx, &organizations[i])
	}

	return organizations
}

// createUserFixtures generates any number of user records for testing. Locations, UserOrganizations, and
// UserAccessTokens are also created for each user. The access token for each user is the same as the user's nickname.
// All user fixtures will be assigned to the first Organization in the DB. If no Organization exists, one will be
// created.
func createUserFixtures(tx *pop.Connection, n int) UserFixtures {
	var org Organization
	if err := tx.First(&org); err != nil {
		org = Organization{AuthConfig: "{}"}
		mustCreate(tx, &org)
	}

	unique := domain.GetUUID().String()

	locations := createLocationFixtures(tx, n)

	users := make(Users, n)
	userOrgs := make(UserOrganizations, n)
	accessTokenFixtures := make(UserAccessTokens, n)
	for i := range users {
		users[i].Email = unique + "_user" + strconv.Itoa(i) + "@example.com"
		users[i].Nickname = unique + "_auth_user" + strconv.Itoa(i)
		users[i].FirstName = "first" + strconv.Itoa(i)
		users[i].LastName = "last" + strconv.Itoa(i)
		users[i].AuthPhotoURL = nulls.NewString("http://example.com/" + users[i].Nickname + ".gif")
		users[i].LocationID = nulls.NewInt(locations[i].ID)
		users[i].AdminRole = UserAdminRoleUser
		mustCreate(tx, &users[i])

		userOrgs[i].UserID = users[i].ID
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].AuthID = users[i].Email
		userOrgs[i].AuthEmail = users[i].Email
		mustCreate(tx, &userOrgs[i])

		if err := tx.Load(&users[i], "Organizations"); err != nil {
			panic(fmt.Sprintf("failed to load organizations on users[%d] fixture, %s", i, err))
		}

		accessTokenFixtures[i].UserID = users[i].ID
		accessTokenFixtures[i].UserOrganizationID = nulls.NewInt(userOrgs[i].ID)
		accessTokenFixtures[i].AccessToken = HashClientIdAccessToken(users[i].Nickname)
		accessTokenFixtures[i].ExpiresAt = time.Now().Add(time.Minute * 60)
		mustCreate(tx, &accessTokenFixtures[i])
	}

	return UserFixtures{
		Organization:      org,
		Users:             users,
		UserOrganizations: userOrgs,
		UserAccessTokens:  accessTokenFixtures,
		Locations:         locations,
	}
}

// createRequestFixtures generates any number of request records for testing. Related Location and File records are also
// created. All request fixtures will be assigned to the first Organization in the DB. If no Organization exists,
// one will be created. All requests are created by the first User in the DB. If no User exists, one will be created.
func createRequestFixtures(tx *pop.Connection, nRequests int, createFiles bool) Requests {
	var org Organization
	if err := tx.First(&org); err != nil {
		org = Organization{AuthConfig: "{}"}
		mustCreate(tx, &org)
	}

	var user User
	if err := tx.First(&user); err != nil {
		user = createUserFixtures(tx, 1).Users[0]
	}

	locations := createLocationFixtures(tx, nRequests*2)

	var files Files
	if createFiles {
		files = createFileFixtures(nRequests)
	}

	requests := make(Requests, nRequests)
	created := 0
	futureDate := time.Now().Add(4 * domain.DurationWeek)
	for i := range requests {
		requests[i].CreatedByID = user.ID
		requests[i].OrganizationID = org.ID
		requests[i].DestinationID = locations[i*2].ID
		requests[i].OriginID = nulls.NewInt(locations[i*2+1].ID)
		requests[i].Title = "title " + strconv.Itoa(i)
		requests[i].Description = nulls.NewString("description " + strconv.Itoa(i))
		requests[i].NeededBefore = nulls.NewTime(futureDate)
		requests[i].Size = RequestSizeSmall
		requests[i].Status = RequestStatusOpen
		requests[i].URL = nulls.NewString("https://www.example.com/" + strconv.Itoa(i))
		requests[i].Kilograms = nulls.NewFloat64(float64(i) * 0.1)
		requests[i].Visibility = RequestVisibilitySame

		if createFiles {
			if _, err := requests[i].AttachPhoto(files[i].UUID.String()); err != nil {
				panic("error attaching photo to request fixture, " + err.Error())
			}
		}

		mustCreate(tx, &requests[i])
		created++
	}

	return requests
}

// createPotentialProviderFixtures generates any number of PotentialProvider records for testing.
// All of these will be assigned to the first Request (Request) in the DB, which has the first User as
// its CreatedBy.
// If necessary, User and Request fixtures will also be created.
func createPotentialProviderFixtures(tx *pop.Connection, nRequests, nProviders int) PotentialProviders {
	var requests Requests
	if err := tx.All(&requests); err != nil {
		createRequestFixtures(tx, nRequests, false)
	}
	if len(requests) < nRequests {
		createRequestFixtures(tx, nRequests-len(requests), false)
	}

	requests = Requests{}
	tx.All(&requests)

	var users Users
	if err := tx.All(&users); err != nil {
		createUserFixtures(tx, nProviders+1)
	}
	if len(users) < nProviders+1 {
		createUserFixtures(tx, nProviders+1-len(users))
	}
	users = Users{}
	tx.All(&users)

	providers := make(PotentialProviders, nProviders)
	for i := range providers {
		providers[i] = PotentialProvider{
			RequestID: requests[0].ID,
			UserID:    users[i+1].ID,
		}
		mustCreate(tx, &providers[i])
	}

	return providers
}

// createLocationFixtures generates any number of location records for testing.
func createLocationFixtures(tx *pop.Connection, n int) Locations {
	countries := []string{"US", "CA", "MX", "TH", "FR", "PG"}
	locations := make(Locations, n)
	for i := range locations {
		locations[i] = Location{
			Country:     countries[rand.Intn(6)],
			Description: "Random Location " + strconv.Itoa(rand.Int()),
			Latitude:    nulls.NewFloat64(rand.Float64()*180 - 90),
			Longitude:   nulls.NewFloat64(rand.Float64()*360 - 180),
		}
		mustCreate(tx, &locations[i])
	}
	return locations
}

func createFileFixtures(n int) Files {
	fileFixtures := make([]File, n)
	for i := range fileFixtures {
		var f File
		if err := f.Store(strconv.Itoa(rand.Int())+".gif", []byte("GIF89a")); err != nil {
			panic(fmt.Sprintf("failed to create file fixture, %s", err))
		}
		fileFixtures[i] = f
	}
	return fileFixtures
}

type potentialProvidersFixtures struct {
	Users
	Requests
	PotentialProviders
}

// createPotentialProviderFixtures generates five PotentialProvider records for testing.
// If necessary, four User and three Request fixtures will also be created.  The Requests will
// all be created by the first user.
// The first Request will have all but the first user as a potential provider.
// The second Request will have the last two users as potential providers.
// The third Request won't have any potential providers
func createPotentialProvidersFixtures(ms *ModelSuite) potentialProvidersFixtures {
	uf := createUserFixtures(ms.DB, 4)
	requests := createRequestFixtures(ms.DB, 3, false)
	providers := PotentialProviders{}

	for i, p := range requests[:2] {
		for _, u := range uf.Users[i+1:] {
			c := PotentialProvider{RequestID: p.ID, UserID: u.ID}
			c.Create()
			providers = append(providers, c)
		}
	}

	return potentialProvidersFixtures{
		Users:              uf.Users,
		Requests:           requests,
		PotentialProviders: providers,
	}
}

// createMeetingFixtures generates any number of meeting records for testing. Related records are also
// created. All meeting fixtures will be assigned to the first Organization in the DB. If no Organization exists,
// one will be created. All meetings are created by the first User in the DB. If no User exists, one will be created.
//
//  Slice index numbers for each object are shown in the following table:
//
//  meeting   invites      participants        organizer user  invited user  self-joined user
//  0         0, 1         0, 1, 2             1               2             3
//  n         n*2, n*2+1   n*3, n*3+1, n*3+2   n*4+1           n*4+2         n*4+3
//
//  Creator for all meetings is user 0
//  Inviter for all invites is user 0
//
//  The first invite is to an existing user
//  The second invite is to a non-user
//
//  The first participant is the meeting organizer
//  The second participant is an invited user
//  The third participant is a self-joined user
func createMeetingFixtures(tx *pop.Connection, nMeetings int) meetingFixtures {
	var org Organization
	if err := tx.First(&org); err != nil {
		org = Organization{AuthConfig: "{}"}
		mustCreate(tx, &org)
	}

	var user User
	if err := tx.First(&user); err != nil {
		user = User{}
		mustCreate(tx, &user)
	}

	locations := createLocationFixtures(tx, nMeetings)

	files := createFileFixtures(nMeetings)

	meetings := make(Meetings, nMeetings)
	for i := range meetings {
		meetings[i].CreatedByID = user.ID
		meetings[i].Name = "meeting " + strconv.Itoa(i)
		meetings[i].LocationID = locations[i].ID
		meetings[i].StartDate = time.Now()
		meetings[i].EndDate = time.Now().Add(time.Hour * 24)
		meetings[i].InviteCode = nulls.NewUUID(domain.GetUUID())
		if _, err := meetings[i].SetImageFile(files[i].UUID.String()); err != nil {
			panic("error attaching image to meeting fixture, " + err.Error())
		}
		mustCreate(tx, &meetings[i])
	}

	const invitesPerMeeting = 2 // one pending and one participating
	const usersPerMeeting = 4   // one organizer + one invited + one invited but not participating + one self-added
	const participantsPerMeeting = usersPerMeeting - 1
	invites := make(MeetingInvites, nMeetings*invitesPerMeeting)
	users := createUserFixtures(tx, nMeetings*usersPerMeeting).Users
	for i := range invites {
		if i%invitesPerMeeting == 0 {
			invites[i].Email = users[i*usersPerMeeting/invitesPerMeeting+1].Email
		}
		if i%invitesPerMeeting == 1 {
			invites[i].Email = users[i*usersPerMeeting/invitesPerMeeting+1].Email
		}
		invites[i].MeetingID = meetings[i/invitesPerMeeting].ID
		invites[i].InviterID = user.ID
		if err := invites[i].Create(); err != nil {
			panic(fmt.Sprintf("error creating invite fixture %d, %s", i, err))
		}
	}

	participants := make(MeetingParticipants, nMeetings*(participantsPerMeeting))
	for i := range participants {
		participants[i].MeetingID = meetings[i/participantsPerMeeting].ID
		participants[i].UserID = users[i*usersPerMeeting/participantsPerMeeting].ID
		if i%participantsPerMeeting == 0 {
			participants[i].IsOrganizer = true
		}
		if i%participantsPerMeeting == 1 {
			participants[i].InviteID = nulls.NewInt(invites[i*invitesPerMeeting/participantsPerMeeting].ID)
		}
		mustCreate(tx, &participants[i])
	}

	return meetingFixtures{
		Meetings:            meetings,
		MeetingInvites:      invites,
		MeetingParticipants: participants,
		Users:               append(Users{user}, users...),
	}
}

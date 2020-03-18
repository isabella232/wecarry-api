package actions

import (
	"fmt"
	"time"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
)

func (as *ActionSuite) Test_AddMeAsPotentialProvider() {

	f := test.CreatePotentialProvidersFixtures(as.DB)
	posts := f.Posts
	const qTemplate = `mutation {post: addMeAsPotentialProvider ` +
		`(input: {postID: "%s" deliveryAfter: "%s" deliveryBefore: "%s"})` +
		` {id title potentialProviders{user{id nickname} deliveryAfter deliveryBefore}}}`

	const qTemplate = `mutation {request: addMeAsPotentialProvider (requestID: "%s")` +
		` {id title potentialProviders{id nickname}}}`
	deliveryAfter := time.Now().Add(domain.DurationWeek).Format(domain.DateFormat)
	deliveryBefore := time.Now().Add(2 * domain.DurationWeek).Format(domain.DateFormat)

	// Add one to Post with none
	query := fmt.Sprintf(qTemplate, posts[2].UUID.String(), deliveryAfter, deliveryBefore)

	var resp RequestResponse

	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)
	as.Equal(posts[2].UUID.String(), resp.Request.ID, "incorrect Post UUID")
	as.Equal(posts[2].Title, resp.Request.Title, "incorrect Post title")

	want := []PotentialProvider{{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname}}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")
	want := []PotentialProvider{
		{User: PublicProfile{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname},
			DeliveryAfter: deliveryAfter, DeliveryBefore: deliveryBefore}}
	as.Equal(want, resp.Post.PotentialProviders, "incorrect potential providers")

	// Add one to Post with two already
	deliveryBefore = time.Now().Add(3 * domain.DurationWeek).Format(domain.DateFormat)
	query = fmt.Sprintf(qTemplate, posts[1].UUID.String(), deliveryAfter, deliveryBefore)

	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)
	as.Equal(posts[1].UUID.String(), resp.Request.ID, "incorrect Post UUID")
	as.Equal(posts[1].Title, resp.Request.Title, "incorrect Post title")

	ppros := f.PotentialProviders
	deliveryAfter0 := ppros[0].DeliveryAfter.Format(domain.DateFormat)
	deliveryBefore0 := ppros[0].DeliveryBefore.Format(domain.DateFormat)
	deliveryAfter1 := ppros[1].DeliveryAfter.Format(domain.DateFormat)
	deliveryBefore1 := ppros[1].DeliveryBefore.Format(domain.DateFormat)

	want = []PotentialProvider{
		{User: PublicProfile{ID: f.Users[2].UUID.String(), Nickname: f.Users[2].Nickname},
			DeliveryAfter: deliveryAfter0, DeliveryBefore: deliveryBefore0},
		{User: PublicProfile{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname},
			DeliveryAfter: deliveryAfter1, DeliveryBefore: deliveryBefore1},
		{User: PublicProfile{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname},
			DeliveryAfter: deliveryAfter, DeliveryBefore: deliveryBefore},
	}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")

	// Adding a repeat gives an error
	query = fmt.Sprintf(qTemplate, posts[1].UUID.String(), deliveryAfter, deliveryBefore)

	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.Error(err, "expected an error (unique together) but didn't get one")

	want = []PotentialProvider{
		{User: PublicProfile{ID: f.Users[2].UUID.String(), Nickname: f.Users[2].Nickname},
			DeliveryAfter: deliveryAfter0, DeliveryBefore: deliveryBefore0},
		{User: PublicProfile{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname},
			DeliveryAfter: deliveryAfter1, DeliveryBefore: deliveryBefore1},
		{User: PublicProfile{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname},
			DeliveryAfter: deliveryAfter, DeliveryBefore: deliveryBefore},
	}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")

	// Adding one for a different Org gives an error
	err = as.testGqlQuery(query, f.Users[4].Nickname, &resp)
	as.Error(err, "expected an error (unauthorized) but didn't get one")
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")

}

func (as *ActionSuite) Test_RemoveMeAsPotentialProvider() {

	f := test.CreatePotentialProvidersFixtures(as.DB)
	posts := f.Posts

	const qTemplate = `mutation {request: removeMeAsPotentialProvider (requestID: "%s")` +
		` {id title potentialProviders{id nickname}}}`
	const qTemplate = `mutation {post: removeMeAsPotentialProvider (postID: "%s")` +
		` {id title potentialProviders{user{id nickname}}}}`

	var resp RequestResponse

	query := fmt.Sprintf(qTemplate, posts[1].UUID.String())

	err := as.testGqlQuery(query, f.Users[2].Nickname, &resp)
	as.NoError(err)
	as.Equal(posts[1].UUID.String(), resp.Request.ID, "incorrect Post UUID")
	as.Equal(posts[1].Title, resp.Request.Title, "incorrect Post title")

	want := []PotentialProvider{{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname}}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")
	want := []PotentialProvider{
		{User: PublicProfile{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname}}}
	as.Equal(want, resp.Post.PotentialProviders, "incorrect potential providers")
}

func (as *ActionSuite) Test_RemovePotentialProvider() {

	f := test.CreatePotentialProvidersFixtures(as.DB)
	posts := f.Posts

	const qTemplate = `mutation {request: removePotentialProvider (requestID: "%s", userID: "%s")` +
		` {id title potentialProviders{id nickname}}}`
	const qTemplate = `mutation {post: removePotentialProvider (postID: "%s", userID: "%s")` +
		` {id title potentialProviders{user{id nickname}}}}`

	var resp RequestResponse

	// remove third User as a potential provider on second Post
	query := fmt.Sprintf(qTemplate, posts[1].UUID.String(), f.Users[2].UUID.String())

	err := as.testGqlQuery(query, f.Users[2].Nickname, &resp)
	as.NoError(err)
	as.Equal(posts[1].UUID.String(), resp.Request.ID, "incorrect Post UUID")
	as.Equal(posts[1].Title, resp.Request.Title, "incorrect Post title")

	want := []PotentialProvider{{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname}}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")
	want := []PotentialProvider{
		{User: PublicProfile{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname}}}
	as.Equal(want, resp.Post.PotentialProviders, "incorrect potential providers")
}

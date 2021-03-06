package actions

import (
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type threadQueryFixtures struct {
	models.Organization
	models.Users
	models.Requests
	models.Threads
	models.Messages
}

func createFixturesForThreadQuery(as *ActionSuite) threadQueryFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	requests := test.CreateRequestFixtures(as.DB, 2, false)

	threads := models.Threads{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
		{UUID: domain.GetUUID(), RequestID: requests[1].ID},
	}
	for i := range threads {
		createFixture(as, &threads[i])
	}

	threadParticipants := models.ThreadParticipants{
		{ThreadID: threads[0].ID, UserID: requests[0].CreatedByID},
	}
	for i := range threadParticipants {
		createFixture(as, &threadParticipants[i])
	}

	messages := models.Messages{
		{
			ThreadID: threads[0].ID,
			SentByID: users[1].ID,
			Content:  "Message from " + users[1].Nickname,
		},
		{
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "Reply from " + users[0].Nickname,
		},
	}
	for i := range messages {
		messages[i].UUID = domain.GetUUID()
		createFixture(as, &messages[i])
	}

	return threadQueryFixtures{
		Organization: org,
		Users:        users,
		Requests:     requests,
		Threads:      threads,
		Messages:     messages,
	}
}

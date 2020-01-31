package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/job"
)

// ServiceInput defines the input parameters to the "service" endpoint
type ServiceInput struct {
	Task ServiceTaskName `json:"task"`
}

// ServiceTask is a type of task to be issued by the "service" endpoint
type ServiceTask struct {
	Handler ServiceTaskHandler
}

// ServiceTaskName is the name of a type of task to be issued by the "service" endpoint
type ServiceTaskName string

// ServiceTaskHandler is a handler function that is executed when a specified task is requested by the API client
type ServiceTaskHandler func(buffalo.Context) error

const (
	// ServiceTaskFileCleanup removes files not linked to any object
	ServiceTaskFileCleanup ServiceTaskName = "file_cleanup"

	// ServiceTaskTokenCleanup removes expired user access tokens
	ServiceTaskTokenCleanup ServiceTaskName = "token_cleanup"
)

var serviceTasks = map[ServiceTaskName]ServiceTask{
	ServiceTaskFileCleanup: {
		Handler: fileCleanupHandler,
	},
	ServiceTaskTokenCleanup: {
		Handler: tokenCleanupHandler,
	},
}

func serviceHandler(c buffalo.Context) error {
	if domain.Env.ServiceIntegrationToken == "" {
		return c.Error(http.StatusInternalServerError, errors.New("no ServiceIntegrationToken configured"))
	}

	bearerToken := domain.GetBearerTokenFromRequest(c.Request())
	if domain.Env.ServiceIntegrationToken != bearerToken {
		return c.Error(http.StatusUnauthorized, errors.New("incorrect bearer token provided"))
	}

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.Error(http.StatusInternalServerError, fmt.Errorf("error reading request body, %s", err))
	}

	var input ServiceInput
	if err := json.Unmarshal(body, &input); err != nil {
		return c.Error(http.StatusBadRequest, fmt.Errorf("error parsing request body, %s", err))
	}

	domain.Logger.Printf("scheduling service task '%s'", input.Task)

	if task, ok := serviceTasks[input.Task]; ok {
		if err := task.Handler(c); err != nil {
			return c.Error(http.StatusInternalServerError, fmt.Errorf("task %s failed, %s", input.Task, err))
		}
		return c.Render(http.StatusNoContent, nil)
	}
	return c.Error(http.StatusUnprocessableEntity, fmt.Errorf("invalid task name: %s", input.Task))
}

func fileCleanupHandler(c buffalo.Context) error {
	if err := job.SubmitDelayed(job.FileCleanup, time.Second, nil); err != nil {
		return c.Error(http.StatusInternalServerError, fmt.Errorf("file cleanup job not started, %s", err))
	}
	return nil
}

func tokenCleanupHandler(c buffalo.Context) error {
	if err := job.SubmitDelayed(job.TokenCleanup, time.Second, nil); err != nil {
		return c.Error(http.StatusInternalServerError, fmt.Errorf("token cleanup job not started, %s", err))
	}
	return nil
}

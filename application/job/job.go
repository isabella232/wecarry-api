package job

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/envy"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

const (
	NewMessage = "new_message"
)

var W worker.Worker

func init() {
	W = worker.NewSimple()
	if err := W.Register(NewMessage, NewMessageHandler); err != nil {
		domain.ErrLogger.Printf("error registering 'new_message' worker, %s", err)
	}
}

// NewMessageHandler is the Worker handler for new notifications of new Thread Messages
func NewMessageHandler(args worker.Args) error {
	domain.Logger.Printf("--------- new_message worker, args: %+v", args)

	id, ok := args[domain.ArgMessageID].(int)
	if !ok {
		return fmt.Errorf("no message ID provided to new_message worker, args = %+v", args)
	}

	var m models.Message
	if err := m.FindByID(id, "SentBy", "Thread"); err != nil {
		return fmt.Errorf("bad ID (%d) received by new message handler, %s", id, err)
	}

	if err := m.Thread.Load("Participants", "Post"); err != nil {
		return errors.New("failed to load Participants and Post in new message handler")
	}

	var recipients []struct{ Nickname, Email string }
	for _, p := range m.Thread.Participants {
		if p.ID == m.SentBy.ID {
			continue
		}

		var tp models.ThreadParticipant
		if err := models.DB.Where("user_id = ? AND thread_id = ?", p.ID, m.ThreadID).First(&tp); err != nil {
			return fmt.Errorf("failed to find thread_participant record for user %d and thread %d, %s",
				tp.ID, m.ThreadID, err)
		}
		// Don't send a notification if this user has viewed the message or if they've already been notified
		if tp.LastViewedAt.After(m.UpdatedAt) || tp.LastNotifiedAt.After(m.UpdatedAt) {
			continue
		}

		tp.LastNotifiedAt = time.Now()
		if err := models.DB.Update(&tp); err != nil {
			return errors.New("failed to update thread_participant.last_notified_at")
		}

		recipients = append(recipients,
			struct{ Nickname, Email string }{p.Nickname, p.Email})
	}

	uiUrl := envy.Get(domain.UIURLEnv, "")
	data := map[string]interface{}{
		"postURL":        uiUrl + "/#/requests/" + m.Thread.Post.Uuid.String(),
		"postTitle":      m.Thread.Post.Title,
		"messageContent": m.Content,
		"sentByNickname": m.SentBy.Nickname,
		"threadURL":      uiUrl + "/#/messages/" + m.Thread.Uuid.String(),
	}

	for _, r := range recipients {
		msg := notifications.Message{
			Template:  domain.MessageTemplateNewMessage,
			Data:      data,
			FromName:  m.SentBy.Nickname,
			FromEmail: m.SentBy.Email,
			ToName:    r.Nickname,
			ToEmail:   r.Email,
		}
		if err := notifications.Send(msg); err != nil {
			domain.ErrLogger.Printf("error sending 'New Message' notification, %s", err)
		}
	}

	return nil
}

// Submit enqueues a new Worker job for the given handler. Arguments can be provided in `args`.
func Submit(handler string, args map[string]interface{}) error {
	job := worker.Job{
		Queue:   "default",
		Args:    args,
		Handler: handler,
	}
	if err := W.PerformIn(job, 10*time.Second); err != nil {
		domain.ErrLogger.Print(err)
		return err
	}

	return nil
}

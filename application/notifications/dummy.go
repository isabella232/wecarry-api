package notifications

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/silinternational/wecarry-api/domain"
)

type DummyEmailService struct {
	sentMessages []dummyMessage
}

var TestEmailService DummyEmailService

type dummyMessage struct {
	subject, body, fromName, fromEmail, toName, toEmail string
}

type dummyTemplate struct {
	subject, body string
}

const providerTag = "_provider"
const receiverTag = "_receiver"

var dummyTemplates = map[string]dummyTemplate{
	domain.MessageTemplateNewRequest: {
		subject: "new request",
		body:    "There is a new request for an item.",
	},
	domain.MessageTemplateNewOffer: {
		subject: "new offer",
		body:    "There is a new offer available.",
	},
	domain.MessageTemplateNewThreadMessage: {
		subject: "new message",
		body:    "You have a new message.",
	},
	domain.MessageTemplateRequestDelivered + receiverTag: {
		subject: domain.MessageTemplateRequestDelivered,
		body:    "The status of a request changed from accepted or committed to delivered.",
	},
	domain.MessageTemplateRequestReceived + providerTag: {
		subject: domain.MessageTemplateRequestReceived,
		body:    "The status of a request changed from accepted or delivered to received.",
	},
	domain.MessageTemplateRequestNotReceivedAfterAll + providerTag: {
		subject: domain.MessageTemplateRequestNotReceivedAfterAll,
		body:    "The status of a request changed from completed to accepted or delivered.",
	},
	domain.MessageTemplateRequestFromAcceptedToCommitted + providerTag: {
		subject: domain.MessageTemplateRequestFromAcceptedToCommitted,
		body:    "The status of a request changed from accepted to committed.",
	},
	domain.MessageTemplateRequestFromAcceptedToDelivered + receiverTag: {
		subject: domain.MessageTemplateRequestFromAcceptedToDelivered,
		body:    "The status of a request changed from accepted to delivered.",
	},
	domain.MessageTemplateRequestFromAcceptedToOpen + providerTag: {
		subject: domain.MessageTemplateRequestFromAcceptedToOpen,
		body:    "The status of a request changed from accepted to open.",
	},
	domain.MessageTemplateRequestFromAcceptedToRemoved + providerTag: {
		subject: domain.MessageTemplateRequestFromAcceptedToRemoved,
		body:    "The status of a request changed from accepted to removed.",
	},
	domain.MessageTemplateRequestFromCommittedToAccepted + providerTag: {
		subject: domain.MessageTemplateRequestFromCommittedToAccepted,
		body:    "The status of a request changed from committed to accepted.",
	},
	domain.MessageTemplateRequestFromCommittedToDelivered + receiverTag: {
		subject: domain.MessageTemplateRequestFromCommittedToDelivered,
		body:    "The status of a request changed from committed to delivered.",
	},
	domain.MessageTemplateRequestFromCommittedToOpen: {
		subject: domain.MessageTemplateRequestFromCommittedToOpen,
		body:    "The status of a request changed from committed to open.",
	},
	domain.MessageTemplateRequestFromCommittedToRemoved + providerTag: {
		subject: domain.MessageTemplateRequestFromCommittedToRemoved,
		body:    "The status of a request changed from committed to removed.",
	},
	domain.MessageTemplateRequestFromDeliveredToAccepted + receiverTag: {
		subject: domain.MessageTemplateRequestFromDeliveredToAccepted,
		body:    "The status of a request changed from delivered to accepted.",
	},
	domain.MessageTemplateRequestFromDeliveredToCommitted + receiverTag: {
		subject: domain.MessageTemplateRequestFromDeliveredToCommitted,
		body:    "The status of a request changed from delivered to committed.",
	},
	domain.MessageTemplateRequestFromDeliveredToCompleted + providerTag: {
		subject: domain.MessageTemplateRequestFromDeliveredToCompleted,
		body:    "The status of a request changed from delivered to completed.",
	},
	domain.MessageTemplateRequestFromOpenToCommitted + receiverTag: {
		subject: domain.MessageTemplateRequestFromOpenToCommitted,
		body:    "The status of a request changed from open to committed.",
	},
	domain.MessageTemplateRequestFromOpenToRemoved: {
		subject: domain.MessageTemplateRequestFromOpenToRemoved,
		body:    "The status of a request changed from open to removed.",
	},
}

func (t *DummyEmailService) Send(msg Message) error {
	_, ok := dummyTemplates[msg.Template]
	if !ok {
		errMsg := fmt.Sprintf("invalid template name: %s", msg.Template)
		domain.ErrLogger.Print(errMsg)
		return errors.New(errMsg)
	}

	eTemplate := msg.Template
	bodyBuf := &bytes.Buffer{}
	if err := eR.HTML(eTemplate).Render(bodyBuf, msg.Data); err != nil {
		errMsg := "error rendering message body - " + err.Error()
		domain.ErrLogger.Print(errMsg)
		return errors.New(errMsg)
	}

	domain.Logger.Printf("dummy message subject: %s, recipient: %s, data: %+v",
		msg.Subject, msg.ToName, msg.Data)

	t.sentMessages = append(t.sentMessages,
		dummyMessage{
			subject:   msg.Subject,
			body:      bodyBuf.String(),
			fromName:  msg.FromName,
			fromEmail: msg.FromEmail,
			toName:    msg.ToName,
			toEmail:   msg.ToEmail,
		})
	return nil
}

// GetNumberOfMessagesSent returns the number of messages sent since initialization or the last call to
// DeleteSentMessages
func (t *DummyEmailService) GetNumberOfMessagesSent() int {
	return len(t.sentMessages)
}

// DeleteSentMessages erases the store of sent messages
func (t *DummyEmailService) DeleteSentMessages() {
	t.sentMessages = []dummyMessage{}
}

func (t *DummyEmailService) GetLastToEmail() string {
	if len(t.sentMessages) == 0 {
		return ""
	}

	return t.sentMessages[len(t.sentMessages)-1].toEmail
}

func (t *DummyEmailService) GetToEmailByIndex(i int) string {
	if len(t.sentMessages) <= i {
		return ""
	}

	return t.sentMessages[i].toEmail
}

func (t *DummyEmailService) GetAllToAddresses() []string {
	emailAddresses := make([]string, len(t.sentMessages))
	for i := range t.sentMessages {
		emailAddresses[i] = t.sentMessages[i].toEmail
	}
	return emailAddresses
}

func (t *DummyEmailService) GetLastBody() string {
	if len(t.sentMessages) == 0 {
		return ""
	}

	return t.sentMessages[len(t.sentMessages)-1].body
}

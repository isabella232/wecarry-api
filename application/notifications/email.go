package notifications

import (
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
	"github.com/silinternational/wecarry-api/domain"
)

var eR = render.New(render.Options{
	HTMLLayout:   "layout.plush.html",
	TemplatesBox: packr.New("app:mailers:templates", "../templates/mail"),
	Helpers:      render.Helpers{},
})

type EmailService interface {
	Send(msg Message) error
}

// GetEmailTemplate returns the filename of the email template corresponding to a particular status change.
//  Most of those will just be the same as the name of the status change.
func GetEmailTemplate(key string) string {
	providerTag := "_provider"
	receiverTag := "_receiver"

	modifiedTemplateNames := map[string]string{
		domain.MessageTemplateRequestFromAcceptedToDelivered:  domain.MessageTemplateRequestDelivered + receiverTag,
		domain.MessageTemplateRequestFromCommittedToDelivered: domain.MessageTemplateRequestDelivered + receiverTag,

		domain.MessageTemplateRequestFromAcceptedToReceived:   domain.MessageTemplateRequestReceived + providerTag,
		domain.MessageTemplateRequestFromAcceptedToCompleted:  domain.MessageTemplateRequestReceived + providerTag,
		domain.MessageTemplateRequestFromDeliveredToCompleted: domain.MessageTemplateRequestReceived + providerTag,
		domain.MessageTemplateRequestFromCommittedToReceived:  domain.MessageTemplateRequestReceived + providerTag,
		domain.MessageTemplateRequestFromCompletedToAccepted:  domain.MessageTemplateRequestNotReceivedAfterAll + providerTag,
		domain.MessageTemplateRequestFromCompletedToDelivered: domain.MessageTemplateRequestNotReceivedAfterAll + providerTag,

		domain.MessageTemplateRequestFromAcceptedToCommitted: domain.MessageTemplateRequestFromAcceptedToCommitted + providerTag,
		domain.MessageTemplateRequestFromAcceptedToOpen:      domain.MessageTemplateRequestFromAcceptedToOpen + providerTag,
		domain.MessageTemplateRequestFromAcceptedToRemoved:   domain.MessageTemplateRequestFromAcceptedToRemoved + providerTag,
		domain.MessageTemplateRequestFromCommittedToAccepted: domain.MessageTemplateRequestFromCommittedToAccepted + providerTag,
		domain.MessageTemplateRequestFromCommittedToRemoved:  domain.MessageTemplateRequestFromCommittedToRemoved + providerTag,

		domain.MessageTemplateRequestFromDeliveredToAccepted:  domain.MessageTemplateRequestFromDeliveredToAccepted + receiverTag,
		domain.MessageTemplateRequestFromDeliveredToCommitted: domain.MessageTemplateRequestFromDeliveredToCommitted + receiverTag,
		domain.MessageTemplateRequestFromOpenToCommitted:      domain.MessageTemplateRequestFromOpenToCommitted + receiverTag,
	}

	template, ok := modifiedTemplateNames[key]
	if !ok {
		template = key
	}

	return template
}

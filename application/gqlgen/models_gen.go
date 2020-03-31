// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlgen

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/silinternational/wecarry-api/models"
)

// Input object for `createMeetingInvites`
type CreateMeetingInvitesInput struct {
	// ID of the `Meeting`
	MeetingID string `json:"meetingID"`
	// Email addresses of the invitees. Duplicate values are ignored.
	Emails []string `json:"emails"`
	// NOT YET IMPLEMENTED -- Send email invites. Default is 'false', do not send any emails.
	SendEmail *bool `json:"sendEmail"`
}

// Input object for `createMeetingParticipant`
type CreateMeetingParticipantInput struct {
	// ID of the `Meeting`
	MeetingID string `json:"meetingID"`
	// Secret code from the `MeetingInvite` or invite code from the `Meeting`. If the `Meeting` is not `INVITE_ONLY`,
	// the code may be omitted.
	Code *string `json:"code"`
	// NOT YET IMPLEMENTED -- Add as a `Meeting` Organizer. Authenticated `User` must be authorized [definition TBD] to do this.
	IsOrganizer *bool `json:"isOrganizer"`
}

type CreateMessageInput struct {
	// message content, limited to 4,096 characters
	Content string `json:"content"`
	// ID of the subject Request
	RequestID string `json:"requestID"`
	// Message thread to which the new message should be attached. If not specified, a new thread is created.
	ThreadID *string `json:"threadID"`
}

type CreateOrganizationDomainInput struct {
	// domain name, limited to 255 characters
	Domain string `json:"domain"`
	// ID of the Organization that owns this domain
	OrganizationID string `json:"organizationID"`
	// Authentication type, overriding the Organization's `authType`. Can be: `saml`, `google`, `azureadv2`.
	AuthType models.AuthType `json:"authType"`
	// Authentication configuration, overriding the Organization's `authConfig. See
	// https://github.com/silinternational/wecarry-api/blob/master/README.md
	AuthConfig *string `json:"authConfig"`
}

type CreateOrganizationInput struct {
	// Organization name, limited to 255 characters
	Name string `json:"name"`
	// Website URL of the Organization, limited to 255 characters
	URL *string `json:"url"`
	// Authentication type for the organization. Can be `saml`, `google`, or `azureadv2`.
	AuthType models.AuthType `json:"authType"`
	// Authentication configuration. See https://github.com/silinternational/wecarry-api/blob/master/README.md
	AuthConfig string `json:"authConfig"`
	// ID of pre-stored image logo file. Upload using the `upload` REST API endpoint.
	LogoFileID *string `json:"logoFileID"`
}

type CreateOrganizationTrustInput struct {
	// ID of one of the two Organizations to join in a trusted affiliation
	PrimaryID string `json:"primaryID"`
	// ID of the second of two Organizations to join in a trusted affiliation
	SecondaryID string `json:"secondaryID"`
}

// Specify a Geographic location
type LocationInput struct {
	// Human-friendly description, e.g. 'Los Angeles, CA, USA'
	Description string `json:"description"`
	// Country (ISO 3166-1 Alpha-2 code), e.g. 'US'
	Country string `json:"country"`
	// Latitude in decimal degrees, e.g. -30.95 = 30 degrees 57 minutes south
	Latitude *float64 `json:"latitude"`
	// Longitude in decimal degrees, e.g. -80.05 = 80 degrees 3 minutes west
	Longitude *float64 `json:"longitude"`
}

type PotentialProviderInput struct {
	RequestID string `json:"requestID"`
	// Date (yyyy-mm-dd) after which the request can be fufilled (NOT inclusive).
	// DeliveryAfter must come before deliveryBefore.
	DeliveryAfter string `json:"deliveryAfter"`
	// Date (yyyy-mm-dd) before which the request can be fulfilled (inclusive, i.e. 'on or before')
	DeliveryBefore string `json:"deliveryBefore"`
}

// User fields that can safely be visible to any user in the system
type PublicProfile struct {
	// unique identifier for the User, the same value as in the `User` type
	ID string `json:"id"`
	// User's nickname. Auto-assigned upon creation of a User, but editable by the User. Limited to 255 characters.
	Nickname string `json:"nickname"`
	// avatarURL is generated from an attached photo if present, an external URL if present, or a Gravatar URL
	AvatarURL *string `json:"avatarURL"`
}

// Input object for `removeMeetingInvite`
type RemoveMeetingInviteInput struct {
	// ID of the `Meeting`
	MeetingID string `json:"meetingID"`
	// Email addresse of the invitee to remove
	Email string `json:"email"`
}

// Input object for `removeMeetingParticipant`
type RemoveMeetingParticipantInput struct {
	// ID of the `Meeting`
	MeetingID string `json:"meetingID"`
	// `User` ID of the `Meeting` participant to remove
	UserID string `json:"userID"`
}

type RemoveOrganizationDomainInput struct {
	// domain name, limited to 255 characters
	Domain string `json:"domain"`
	// ID of the Organization that owns this domain
	OrganizationID string `json:"organizationID"`
}

type RemoveOrganizationTrustInput struct {
	// ID of one of the two Organizations in the trust to be removed
	PrimaryID string `json:"primaryID"`
	// ID of the second of two Organizations in the trust to be removed
	SecondaryID string `json:"secondaryID"`
}

type RemoveWatchInput struct {
	// unique identifier for the Watch to be removed
	ID string `json:"id"`
}

type SetThreadLastViewedAtInput struct {
	ThreadID string    `json:"threadID"`
	Time     time.Time `json:"time"`
}

type UpdateOrganizationInput struct {
	// unique identifier for the Organization to be updated
	ID string `json:"id"`
	// Organization name, limited to 255 characters
	Name string `json:"name"`
	// Website URL of the Organization, limited to 255 characters. If omitted, existing URL is erased.
	URL *string `json:"url"`
	// Authentication type for the organization. Can be 'saml', 'google', or 'azureadv2'.
	AuthType models.AuthType `json:"authType"`
	// Authentication configuration. See https://github.com/silinternational/wecarry-api/blob/master/README.md
	AuthConfig string `json:"authConfig"`
	// ID of image logo file. Upload using the `upload` REST API endpoint. If omitted, existing logo is erased.
	LogoFileID *string `json:"logoFileID"`
}

type UpdateRequestStatusInput struct {
	// ID of the request to update
	ID string `json:"id"`
	// New Status. Only a limited set of transitions are allowed.
	Status models.RequestStatus `json:"status"`
	// User ID of the accepted provider. Required if `status` is ACCEPTED and ignored otherwise.
	ProviderUserID *string `json:"providerUserID"`
}

// Input object for `updateUser`
type UpdateUserInput struct {
	// unique identifier for the User to be updated
	ID *string `json:"id"`
	// User's nickname. Auto-assigned upon creation of a User, but editable by the User. Limited to 255 characters.
	Nickname *string `json:"nickname"`
	// File ID of avatar photo. If omitted or `null`, the photo is removed from the profile.
	PhotoID *string `json:"photoID"`
	// Specify the user's 'home' location. If omitted or `null`, the location is removed from the profile.
	Location *LocationInput `json:"location"`
	// New user preferences. If `null` no changes are made.
	Preferences *UpdateUserPreferencesInput `json:"preferences"`
}

type UpdateUserPreferencesInput struct {
	// preferred language -- if omitted, the preference is set to the App default
	Language *PreferredLanguage `json:"language"`
	// time zone -- if omitted, the preference is set to the App default
	TimeZone *string `json:"timeZone"`
	// weight unit-- if omitted, the preference is set to the App default
	WeightUnit *PreferredWeightUnit `json:"weightUnit"`
}

// The User who has offered to fufill the request and the delivery date range
type PotentialProvider struct {
	User *PublicProfile `json:"user"`
	// Date (yyyy-mm-dd). DeliveryAfter is NOT intended to be inclusive of the day itself.
	DeliveryAfter string `json:"deliveryAfter"`
	// Date (yyyy-mm-dd). It is intended to be inclusive of the day itself (i.e. 'on or before')
	DeliveryBefore string `json:"deliveryBefore"`
}

// Visibility for Meetings (Events), determines who can see a `Meeting`.
type MeetingVisibility string

const (
	// Visible to invitees and all app users
	MeetingVisibilityAll MeetingVisibility = "ALL"
	// Visible to invitees and members of the `Meeting` organization and affiliated organizations
	MeetingVisibilityTrusted MeetingVisibility = "TRUSTED"
	// Visible to invitees and members of the `Meeting` organization
	MeetingVisibilityOrganization MeetingVisibility = "ORGANIZATION"
	// Visible only to invitees
	MeetingVisibilityInviteOnly MeetingVisibility = "INVITE_ONLY"
)

var AllMeetingVisibility = []MeetingVisibility{
	MeetingVisibilityAll,
	MeetingVisibilityTrusted,
	MeetingVisibilityOrganization,
	MeetingVisibilityInviteOnly,
}

func (e MeetingVisibility) IsValid() bool {
	switch e {
	case MeetingVisibilityAll, MeetingVisibilityTrusted, MeetingVisibilityOrganization, MeetingVisibilityInviteOnly:
		return true
	}
	return false
}

func (e MeetingVisibility) String() string {
	return string(e)
}

func (e *MeetingVisibility) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = MeetingVisibility(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid MeetingVisibility", str)
	}
	return nil
}

func (e MeetingVisibility) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// User's preferred language, used for translation of system text messages. (ISO 639-1 code)
type PreferredLanguage string

const (
	// English
	PreferredLanguageEn PreferredLanguage = "EN"
	// French
	PreferredLanguageFr PreferredLanguage = "FR"
	// Spanish
	PreferredLanguageEs PreferredLanguage = "ES"
	// Korean
	PreferredLanguageKo PreferredLanguage = "KO"
	// Portuguese
	PreferredLanguagePt PreferredLanguage = "PT"
)

var AllPreferredLanguage = []PreferredLanguage{
	PreferredLanguageEn,
	PreferredLanguageFr,
	PreferredLanguageEs,
	PreferredLanguageKo,
	PreferredLanguagePt,
}

func (e PreferredLanguage) IsValid() bool {
	switch e {
	case PreferredLanguageEn, PreferredLanguageFr, PreferredLanguageEs, PreferredLanguageKo, PreferredLanguagePt:
		return true
	}
	return false
}

func (e PreferredLanguage) String() string {
	return string(e)
}

func (e *PreferredLanguage) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PreferredLanguage(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PreferredLanguage", str)
	}
	return nil
}

func (e PreferredLanguage) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// User's preferred weight units
type PreferredWeightUnit string

const (
	PreferredWeightUnitPounds    PreferredWeightUnit = "POUNDS"
	PreferredWeightUnitKilograms PreferredWeightUnit = "KILOGRAMS"
)

var AllPreferredWeightUnit = []PreferredWeightUnit{
	PreferredWeightUnitPounds,
	PreferredWeightUnitKilograms,
}

func (e PreferredWeightUnit) IsValid() bool {
	switch e {
	case PreferredWeightUnitPounds, PreferredWeightUnitKilograms:
		return true
	}
	return false
}

func (e PreferredWeightUnit) String() string {
	return string(e)
}

func (e *PreferredWeightUnit) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PreferredWeightUnit(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PreferredWeightUnit", str)
	}
	return nil
}

func (e PreferredWeightUnit) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// Context of a User with respect to a Request
type RequestRole string

const (
	// Requests created by the User
	RequestRoleCreatedby RequestRole = "CREATEDBY"
	// Requests provided by the User. Requests where the user is a PotentialProvider are not included.
	RequestRoleProviding RequestRole = "PROVIDING"
)

var AllRequestRole = []RequestRole{
	RequestRoleCreatedby,
	RequestRoleProviding,
}

func (e RequestRole) IsValid() bool {
	switch e {
	case RequestRoleCreatedby, RequestRoleProviding:
		return true
	}
	return false
}

func (e RequestRole) String() string {
	return string(e)
}

func (e *RequestRole) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RequestRole(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RequestRole", str)
	}
	return nil
}

func (e RequestRole) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

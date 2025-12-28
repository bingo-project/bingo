// ABOUTME: Notification channel constants.
// ABOUTME: Defines in-app, email, SMS, and push channels.

package notification

type Channel string

const (
	ChannelInApp Channel = "in_app"
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"  // Reserved
	ChannelPush  Channel = "push" // Reserved
)

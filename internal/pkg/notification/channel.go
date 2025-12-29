// ABOUTME: Notification channel constants.
// ABOUTME: Defines delivery channels and Redis Pub/Sub channel names.

package notification

type Channel string

const (
	ChannelInApp Channel = "in_app"
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"  // Reserved
	ChannelPush  Channel = "push" // Reserved
)

// Redis Pub/Sub channel names.
const (
	RedisBroadcastChannel  = "ntf:broadcast"
	RedisUserChannelPrefix = "ntf:user:"
)

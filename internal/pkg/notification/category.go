// ABOUTME: Notification category constants.
// ABOUTME: Defines system, security, transaction, and social categories.

package notification

type Category string

const (
	CategorySystem      Category = "system"
	CategorySecurity    Category = "security"
	CategoryTransaction Category = "transaction"
	CategorySocial      Category = "social"
)

// ABOUTME: AI quota model definitions.
// ABOUTME: Defines quota tiers and per-user quota tracking.

package model

import "time"

// AiQuotaTierM represents quota tier definitions.
type AiQuotaTierM struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Tier        string `gorm:"column:tier;type:varchar(32);uniqueIndex:uk_tier;not null" json:"tier"`
	DisplayName string `gorm:"column:display_name;type:varchar(64)" json:"displayName"`
	RPM         int    `gorm:"column:rpm;type:int;not null;default:10" json:"rpm"`
	TPD         int    `gorm:"column:tpd;type:int;not null;default:100000" json:"tpd"`

	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (*AiQuotaTierM) TableName() string {
	return "ai_quota_tier"
}

// AiUserQuotaM represents per-user quota tracking.
type AiUserQuotaM struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UID             string     `gorm:"column:uid;type:varchar(64);uniqueIndex:uk_uid;not null" json:"uid"`
	Tier            string     `gorm:"column:tier;type:varchar(32);index:idx_tier;not null;default:free" json:"tier"`
	RPM             int        `gorm:"column:rpm;type:int;not null;default:0" json:"rpm"`
	TPD             int        `gorm:"column:tpd;type:int;not null;default:0" json:"tpd"`
	UsedTokensToday int        `gorm:"column:used_tokens_today;type:int;not null;default:0" json:"usedTokensToday"`
	LastResetAt     *time.Time `gorm:"column:last_reset_at;type:timestamp;default:null" json:"lastResetAt"`

	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (*AiUserQuotaM) TableName() string {
	return "ai_user_quota"
}

// Quota tier constants.
const (
	AiQuotaTierFree       = "free"
	AiQuotaTierPro        = "pro"
	AiQuotaTierEnterprise = "enterprise"
)

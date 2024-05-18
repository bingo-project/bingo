package model

type UserAccount struct {
	Base

	UID       string `gorm:"type:varchar(255);index:idx_uid"`
	Provider  string `gorm:"type:varchar(255);not null;default:''"`
	AccountID string `gorm:"type:varchar(255);not null;default:''"`
	Username  string `gorm:"type:varchar(255);not null;default:''"`
	Nickname  string `gorm:"type:varchar(255);not null;default:''"`
	Email     string `gorm:"type:varchar(255);not null;default:''"`
	Bio       string `gorm:"type:varchar(255);not null;default:''"`
	Avatar    string `gorm:"type:varchar(255);not null;default:''"`
}

func (*UserAccount) TableName() string {
	return "uc_user_account"
}

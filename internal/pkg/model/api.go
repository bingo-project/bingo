package model

type ApiM struct {
	Base

	Method      string `gorm:"type:varchar(255);not null;default:''"`
	Path        string `gorm:"type:varchar(255);not null;default:''"`
	Group       string `gorm:"type:varchar(255);not null;default:''"`
	Description string `gorm:"type:varchar(255);not null;default:''"`
}

func (u *ApiM) TableName() string {
	return "sys_auth_api"
}

type Apis []ApiM

func (arr Apis) Len() int {
	return len(arr)
}

func (arr Apis) Less(i, j int) bool {
	return arr[i].Path < arr[j].Path
}

func (arr Apis) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

package biz

//go:generate mockgen -destination mock_biz.go -package biz bingo/internal/apiserver/biz IBiz

import (
	"bingo/internal/apiserver/biz/common"
	"bingo/internal/apiserver/biz/file"
	"bingo/internal/apiserver/biz/syscfg"
	"bingo/internal/apiserver/biz/system"
	"bingo/internal/apiserver/biz/user"
	"bingo/internal/apiserver/store"
)

// IBiz 定义了 Biz 层需要实现的方法.
type IBiz interface {
	Users() user.UserBiz
	Admins() system.AdminBiz
	Roles() system.RoleBiz
	Apis() system.ApiBiz
	Menus() system.MenuBiz
	Files() file.FileBiz
	Email() common.EmailBiz
	Apps() syscfg.AppBiz
	Configs() syscfg.ConfigBiz
}

// biz 是 IBiz 的一个具体实现.
type biz struct {
	ds store.IStore
}

// 确保 biz 实现了 IBiz 接口.
var _ IBiz = (*biz)(nil)

// NewBiz 创建一个 IBiz 类型的实例.
func NewBiz(ds store.IStore) *biz {
	return &biz{ds: ds}
}

// Users 返回一个实现了 UserBiz 接口的实例.
func (b *biz) Users() user.UserBiz {
	return user.New(b.ds)
}

// Admins 管理员.
func (b *biz) Admins() system.AdminBiz {
	return system.NewAdmin(b.ds)
}

// Roles 角色管理.
func (b *biz) Roles() system.RoleBiz {
	return system.NewRole(b.ds)
}

func (b *biz) Apis() system.ApiBiz {
	return system.NewApi(b.ds)
}

func (b *biz) Menus() system.MenuBiz {
	return system.NewMenu(b.ds)
}

func (b *biz) Files() file.FileBiz {
	return file.NewFile(b.ds)
}

func (b *biz) Email() common.EmailBiz {
	return common.NewEmail(b.ds)
}

func (b *biz) Apps() syscfg.AppBiz {
	return syscfg.NewApp(b.ds)
}

func (b *biz) Configs() syscfg.ConfigBiz {
	return syscfg.NewConfig(b.ds)
}

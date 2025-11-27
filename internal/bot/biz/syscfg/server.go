package syscfg

import (
	"context"

	"bingo/internal/pkg/store"
	model "bingo/internal/pkg/model/syscfg"
)

type ServerBiz interface {
	Status(ctx context.Context) (status string, err error)
	ToggleMaintenance(ctx context.Context) error
}

type serverBiz struct {
	ds store.IStore
}

var _ ServerBiz = (*serverBiz)(nil)

func NewServer(ds store.IStore) *serverBiz {
	return &serverBiz{ds: ds}
}

func (b *serverBiz) Status(ctx context.Context) (status string, err error) {
	server, err := b.ds.SysConfig().GetServerConfig(ctx)
	if err != nil {
		return
	}

	return string(server.Status), nil
}

func (b *serverBiz) ToggleMaintenance(ctx context.Context) error {
	server, err := b.ds.SysConfig().GetServerConfig(ctx)
	if err != nil {
		return err
	}

	toggle := model.ServerStatusMaintenance
	if server.Status == model.ServerStatusMaintenance {
		toggle = model.ServerStatusOK
	}

	server.Status = toggle

	return b.ds.SysConfig().UpdateServerConfig(ctx, server)
}

package syscfg

import (
	"context"

	model "bingo/internal/apiserver/model/syscfg"
	"bingo/internal/apiserver/store"
)

type ServerBiz interface {
	ToggleMaintenance(ctx context.Context) error
}

type serverBiz struct {
	ds store.IStore
}

var _ ServerBiz = (*serverBiz)(nil)

func NewServer(ds store.IStore) *serverBiz {
	return &serverBiz{ds: ds}
}

func (b *serverBiz) ToggleMaintenance(ctx context.Context) error {
	server, err := b.ds.Configs().GetServerConfig(ctx)
	if err != nil {
		return err
	}

	toggle := model.ServerStatusMaintenance
	if server.Status == model.ServerStatusMaintenance {
		toggle = model.ServerStatusOK
	}

	server.Status = toggle

	return b.ds.Configs().UpdateServerConfig(ctx, server)
}

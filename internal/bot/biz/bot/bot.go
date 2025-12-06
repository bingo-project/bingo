package bot

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"

	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	model "bingo/internal/pkg/model/bot"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1/bot"
	"bingo/pkg/store/where"
)

type BotBiz interface {
	List(ctx context.Context, req *v1.ListBotRequest) (*v1.ListBotResponse, error)
	Create(ctx context.Context, req *v1.CreateBotRequest) (*v1.BotInfo, error)
	Get(ctx context.Context, ID uint) (*v1.BotInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdateBotRequest) (*v1.BotInfo, error)
	Delete(ctx context.Context, ID uint) error
}

type botBiz struct {
	ds store.IStore
}

var _ BotBiz = (*botBiz)(nil)

func NewBot(ds store.IStore) *botBiz {
	return &botBiz{ds: ds}
}

func (b *botBiz) List(ctx context.Context, req *v1.ListBotRequest) (*v1.ListBotResponse, error) {
	count, list, err := b.ds.Bot().ListWithRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list bots", "err", err)

		return nil, err
	}

	data := make([]v1.BotInfo, 0)
	for _, item := range list {
		var bot v1.BotInfo
		_ = copier.Copy(&bot, item)

		data = append(data, bot)
	}

	return &v1.ListBotResponse{Total: count, Data: data}, nil
}

func (b *botBiz) Create(ctx context.Context, req *v1.CreateBotRequest) (*v1.BotInfo, error) {
	var botM model.Bot
	_ = copier.Copy(&botM, req)

	err := b.ds.Bot().Create(ctx, &botM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.BotInfo
	_ = copier.Copy(&resp, botM)

	return &resp, nil
}

func (b *botBiz) Get(ctx context.Context, ID uint) (*v1.BotInfo, error) {
	bot, err := b.ds.Bot().Get(ctx, where.F("id", ID))
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	var resp v1.BotInfo
	_ = copier.Copy(&resp, bot)

	return &resp, nil
}

func (b *botBiz) Update(ctx context.Context, ID uint, req *v1.UpdateBotRequest) (*v1.BotInfo, error) {
	botM, err := b.ds.Bot().Get(ctx, where.F("id", ID))
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	if req.Name != nil {
		botM.Name = *req.Name
	}
	if req.Source != nil {
		botM.Source = model.Source(*req.Source)
	}
	if req.Description != nil {
		botM.Description = *req.Description
	}
	if req.Token != nil {
		botM.Token = *req.Token
	}
	if req.Enabled != nil {
		botM.Enabled = *req.Enabled
	}

	if err := b.ds.Bot().Update(ctx, botM); err != nil {
		return nil, err
	}

	var resp v1.BotInfo
	_ = copier.Copy(&resp, botM)

	return &resp, nil
}

func (b *botBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.Bot().Delete(ctx, where.F("id", ID))
}

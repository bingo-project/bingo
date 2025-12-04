package bot

import (
	"context"
	"regexp"

	"github.com/bingo-project/component-base/log"
	"github.com/jinzhu/copier"

	"bingo/internal/pkg/store"
	"bingo/internal/pkg/errno"
	model "bingo/internal/pkg/model/bot"
	v1 "bingo/pkg/api/apiserver/v1/bot"
)

type ChannelBiz interface {
	List(ctx context.Context, req *v1.ListChannelRequest) (*v1.ListChannelResponse, error)
	Create(ctx context.Context, req *v1.CreateChannelRequest) (*v1.ChannelInfo, error)
	Get(ctx context.Context, ID uint) (*v1.ChannelInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdateChannelRequest) (*v1.ChannelInfo, error)
	Delete(ctx context.Context, ID uint) error

	DeleteChannel(ctx context.Context, channelID string) error
}

type channelBiz struct {
	ds store.IStore
}

var _ ChannelBiz = (*channelBiz)(nil)

func NewChannel(ds store.IStore) *channelBiz {
	return &channelBiz{ds: ds}
}

func (b *channelBiz) List(ctx context.Context, req *v1.ListChannelRequest) (*v1.ListChannelResponse, error) {
	count, list, err := b.ds.BotChannel().ListWithRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list channels", "err", err)

		return nil, err
	}

	data := make([]v1.ChannelInfo, 0)
	for _, item := range list {
		var channel v1.ChannelInfo
		_ = copier.Copy(&channel, item)

		data = append(data, channel)
	}

	return &v1.ListChannelResponse{Total: count, Data: data}, nil
}

func (b *channelBiz) Create(ctx context.Context, req *v1.CreateChannelRequest) (*v1.ChannelInfo, error) {
	var channelM model.Channel
	_ = copier.Copy(&channelM, req)

	err := b.ds.BotChannel().Create(ctx, &channelM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.ChannelInfo
	_ = copier.Copy(&resp, channelM)

	return &resp, nil
}

func (b *channelBiz) Get(ctx context.Context, ID uint) (*v1.ChannelInfo, error) {
	channel, err := b.ds.BotChannel().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	var resp v1.ChannelInfo
	_ = copier.Copy(&resp, channel)

	return &resp, nil
}

func (b *channelBiz) Update(ctx context.Context, ID uint, req *v1.UpdateChannelRequest) (*v1.ChannelInfo, error) {
	channelM, err := b.ds.BotChannel().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	if req.Source != nil {
		channelM.Source = model.Source(*req.Source)
	}
	if req.ChannelID != nil {
		channelM.ChannelID = *req.ChannelID
	}
	if req.Author != nil {
		channelM.Author = *req.Author
	}

	if err := b.ds.BotChannel().Update(ctx, channelM); err != nil {
		return nil, err
	}

	var resp v1.ChannelInfo
	_ = copier.Copy(&resp, channelM)

	return &resp, nil
}

func (b *channelBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.BotChannel().DeleteByID(ctx, ID)
}

func (b *channelBiz) DeleteChannel(ctx context.Context, channelID string) error {
	return b.ds.BotChannel().DeleteChannel(ctx, channelID)
}

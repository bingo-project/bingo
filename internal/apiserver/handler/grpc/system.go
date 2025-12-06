// ABOUTME: gRPC system method handlers.
// ABOUTME: Provides healthz and version endpoints for gRPC clients.

package grpc

import (
	"context"
	"net"

	"github.com/bingo-project/component-base/version"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/timestamppb"

	"bingo/internal/pkg/log"
	v1 "bingo/pkg/proto/apiserver/v1/pb"
)

func (h *Handler) Healthz(ctx context.Context, req *v1.HealthzRequest) (*v1.HealthzReply, error) {
	log.C(ctx).Infow("Healthz function called.")

	ret := &v1.HealthzReply{
		Status: "OK",
		Ip:     GetPeerAddr(ctx),
		Ts:     timestamppb.Now(),
	}

	return ret, nil
}

func (h *Handler) Version(ctx context.Context, req *v1.VersionRequest) (*v1.VersionReply, error) {
	log.C(ctx).Infow("Version function called.")

	v := version.Get()

	var ret v1.VersionReply
	_ = copier.Copy(&ret, v)

	return &ret, nil
}

func GetPeerAddr(ctx context.Context) string {
	var addr string
	if pr, ok := peer.FromContext(ctx); ok {
		if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
			addr = tcpAddr.IP.String()
		} else {
			addr = pr.Addr.String()
		}
	}

	return addr
}

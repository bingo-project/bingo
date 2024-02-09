// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.1
// source: apiserver.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ApiServerClient is the client API for ApiServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ApiServerClient interface {
	Healthz(ctx context.Context, in *HealthzRequest, opts ...grpc.CallOption) (*HealthzReply, error)
	Version(ctx context.Context, in *VersionRequest, opts ...grpc.CallOption) (*VersionReply, error)
}

type apiServerClient struct {
	cc grpc.ClientConnInterface
}

func NewApiServerClient(cc grpc.ClientConnInterface) ApiServerClient {
	return &apiServerClient{cc}
}

func (c *apiServerClient) Healthz(ctx context.Context, in *HealthzRequest, opts ...grpc.CallOption) (*HealthzReply, error) {
	out := new(HealthzReply)
	err := c.cc.Invoke(ctx, "/v1.ApiServer/Healthz", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiServerClient) Version(ctx context.Context, in *VersionRequest, opts ...grpc.CallOption) (*VersionReply, error) {
	out := new(VersionReply)
	err := c.cc.Invoke(ctx, "/v1.ApiServer/Version", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ApiServerServer is the server API for ApiServer service.
// All implementations must embed UnimplementedApiServerServer
// for forward compatibility
type ApiServerServer interface {
	Healthz(context.Context, *HealthzRequest) (*HealthzReply, error)
	Version(context.Context, *VersionRequest) (*VersionReply, error)
	mustEmbedUnimplementedApiServerServer()
}

// UnimplementedApiServerServer must be embedded to have forward compatible implementations.
type UnimplementedApiServerServer struct {
}

func (UnimplementedApiServerServer) Healthz(context.Context, *HealthzRequest) (*HealthzReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Healthz not implemented")
}
func (UnimplementedApiServerServer) Version(context.Context, *VersionRequest) (*VersionReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Version not implemented")
}
func (UnimplementedApiServerServer) mustEmbedUnimplementedApiServerServer() {}

// UnsafeApiServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ApiServerServer will
// result in compilation errors.
type UnsafeApiServerServer interface {
	mustEmbedUnimplementedApiServerServer()
}

func RegisterApiServerServer(s grpc.ServiceRegistrar, srv ApiServerServer) {
	s.RegisterService(&ApiServer_ServiceDesc, srv)
}

func _ApiServer_Healthz_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthzRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServerServer).Healthz(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.ApiServer/Healthz",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServerServer).Healthz(ctx, req.(*HealthzRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiServer_Version_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiServerServer).Version(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.ApiServer/Version",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiServerServer).Version(ctx, req.(*VersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ApiServer_ServiceDesc is the grpc.ServiceDesc for ApiServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ApiServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "v1.ApiServer",
	HandlerType: (*ApiServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Healthz",
			Handler:    _ApiServer_Healthz_Handler,
		},
		{
			MethodName: "Version",
			Handler:    _ApiServer_Version_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "apiserver.proto",
}
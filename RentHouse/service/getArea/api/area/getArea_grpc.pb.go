// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.1
// source: area/getArea.proto

package getArea

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

// GetAreaClient is the client API for GetArea service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GetAreaClient interface {
	GetAreaSer(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type getAreaClient struct {
	cc grpc.ClientConnInterface
}

func NewGetAreaClient(cc grpc.ClientConnInterface) GetAreaClient {
	return &getAreaClient{cc}
}

func (c *getAreaClient) GetAreaSer(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/getArea.GetArea/GetAreaSer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetAreaServer is the server API for GetArea service.
// All implementations must embed UnimplementedGetAreaServer
// for forward compatibility
type GetAreaServer interface {
	GetAreaSer(context.Context, *Request) (*Response, error)
	mustEmbedUnimplementedGetAreaServer()
}

// UnimplementedGetAreaServer must be embedded to have forward compatible implementations.
type UnimplementedGetAreaServer struct {
}

func (UnimplementedGetAreaServer) GetAreaSer(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAreaSer not implemented")
}
func (UnimplementedGetAreaServer) mustEmbedUnimplementedGetAreaServer() {}

// UnsafeGetAreaServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GetAreaServer will
// result in compilation errors.
type UnsafeGetAreaServer interface {
	mustEmbedUnimplementedGetAreaServer()
}

func RegisterGetAreaServer(s grpc.ServiceRegistrar, srv GetAreaServer) {
	s.RegisterService(&GetArea_ServiceDesc, srv)
}

func _GetArea_GetAreaSer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GetAreaServer).GetAreaSer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/getArea.GetArea/GetAreaSer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GetAreaServer).GetAreaSer(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

// GetArea_ServiceDesc is the grpc.ServiceDesc for GetArea service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GetArea_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "getArea.GetArea",
	HandlerType: (*GetAreaServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAreaSer",
			Handler:    _GetArea_GetAreaSer_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "area/getArea.proto",
}

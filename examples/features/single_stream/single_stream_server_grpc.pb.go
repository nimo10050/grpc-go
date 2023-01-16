// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.0
// source: single_stream_server.proto

package singlestream

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

// SingleStreamServerClient is the client API for SingleStreamServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SingleStreamServerClient interface {
	LotsOfReplies(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (SingleStreamServer_LotsOfRepliesClient, error)
}

type singleStreamServerClient struct {
	cc grpc.ClientConnInterface
}

func NewSingleStreamServerClient(cc grpc.ClientConnInterface) SingleStreamServerClient {
	return &singleStreamServerClient{cc}
}

func (c *singleStreamServerClient) LotsOfReplies(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (SingleStreamServer_LotsOfRepliesClient, error) {
	stream, err := c.cc.NewStream(ctx, &SingleStreamServer_ServiceDesc.Streams[0], "/single_stream.SingleStreamServer/LotsOfReplies", opts...)
	if err != nil {
		return nil, err
	}
	x := &singleStreamServerLotsOfRepliesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SingleStreamServer_LotsOfRepliesClient interface {
	Recv() (*StreamResponse, error)
	grpc.ClientStream
}

type singleStreamServerLotsOfRepliesClient struct {
	grpc.ClientStream
}

func (x *singleStreamServerLotsOfRepliesClient) Recv() (*StreamResponse, error) {
	m := new(StreamResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SingleStreamServerServer is the server API for SingleStreamServer service.
// All implementations must embed UnimplementedSingleStreamServerServer
// for forward compatibility
type SingleStreamServerServer interface {
	LotsOfReplies(*StreamRequest, SingleStreamServer_LotsOfRepliesServer) error
	mustEmbedUnimplementedSingleStreamServerServer()
}

// UnimplementedSingleStreamServerServer must be embedded to have forward compatible implementations.
type UnimplementedSingleStreamServerServer struct {
}

func (UnimplementedSingleStreamServerServer) LotsOfReplies(*StreamRequest, SingleStreamServer_LotsOfRepliesServer) error {
	return status.Errorf(codes.Unimplemented, "method LotsOfReplies not implemented")
}
func (UnimplementedSingleStreamServerServer) mustEmbedUnimplementedSingleStreamServerServer() {}

// UnsafeSingleStreamServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SingleStreamServerServer will
// result in compilation errors.
type UnsafeSingleStreamServerServer interface {
	mustEmbedUnimplementedSingleStreamServerServer()
}

func RegisterSingleStreamServerServer(s grpc.ServiceRegistrar, srv SingleStreamServerServer) {
	s.RegisterService(&SingleStreamServer_ServiceDesc, srv)
}

func _SingleStreamServer_LotsOfReplies_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SingleStreamServerServer).LotsOfReplies(m, &singleStreamServerLotsOfRepliesServer{stream})
}

type SingleStreamServer_LotsOfRepliesServer interface {
	Send(*StreamResponse) error
	grpc.ServerStream
}

type singleStreamServerLotsOfRepliesServer struct {
	grpc.ServerStream
}

func (x *singleStreamServerLotsOfRepliesServer) Send(m *StreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

// SingleStreamServer_ServiceDesc is the grpc.ServiceDesc for SingleStreamServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SingleStreamServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "single_stream.SingleStreamServer",
	HandlerType: (*SingleStreamServerServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "LotsOfReplies",
			Handler:       _SingleStreamServer_LotsOfReplies_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "single_stream_server.proto",
}
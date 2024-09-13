// Since: cosmos-sdk 0.43

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: cosmos/feegrant/v1beta1/tx.proto

package feegrantv1beta1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Msg_GrantAllowance_FullMethodName  = "/cosmos.feegrant.v1beta1.Msg/GrantAllowance"
	Msg_RevokeAllowance_FullMethodName = "/cosmos.feegrant.v1beta1.Msg/RevokeAllowance"
	Msg_PruneAllowances_FullMethodName = "/cosmos.feegrant.v1beta1.Msg/PruneAllowances"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Msg defines the feegrant msg service.
type MsgClient interface {
	// GrantAllowance grants fee allowance to the grantee on the granter's
	// account with the provided expiration time.
	GrantAllowance(ctx context.Context, in *MsgGrantAllowance, opts ...grpc.CallOption) (*MsgGrantAllowanceResponse, error)
	// RevokeAllowance revokes any fee allowance of granter's account that
	// has been granted to the grantee.
	RevokeAllowance(ctx context.Context, in *MsgRevokeAllowance, opts ...grpc.CallOption) (*MsgRevokeAllowanceResponse, error)
	// PruneAllowances prunes expired fee allowances, currently up to 75 at a time.
	PruneAllowances(ctx context.Context, in *MsgPruneAllowances, opts ...grpc.CallOption) (*MsgPruneAllowancesResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) GrantAllowance(ctx context.Context, in *MsgGrantAllowance, opts ...grpc.CallOption) (*MsgGrantAllowanceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgGrantAllowanceResponse)
	err := c.cc.Invoke(ctx, Msg_GrantAllowance_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RevokeAllowance(ctx context.Context, in *MsgRevokeAllowance, opts ...grpc.CallOption) (*MsgRevokeAllowanceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgRevokeAllowanceResponse)
	err := c.cc.Invoke(ctx, Msg_RevokeAllowance_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) PruneAllowances(ctx context.Context, in *MsgPruneAllowances, opts ...grpc.CallOption) (*MsgPruneAllowancesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgPruneAllowancesResponse)
	err := c.cc.Invoke(ctx, Msg_PruneAllowances_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility.
//
// Msg defines the feegrant msg service.
type MsgServer interface {
	// GrantAllowance grants fee allowance to the grantee on the granter's
	// account with the provided expiration time.
	GrantAllowance(context.Context, *MsgGrantAllowance) (*MsgGrantAllowanceResponse, error)
	// RevokeAllowance revokes any fee allowance of granter's account that
	// has been granted to the grantee.
	RevokeAllowance(context.Context, *MsgRevokeAllowance) (*MsgRevokeAllowanceResponse, error)
	// PruneAllowances prunes expired fee allowances, currently up to 75 at a time.
	PruneAllowances(context.Context, *MsgPruneAllowances) (*MsgPruneAllowancesResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) GrantAllowance(context.Context, *MsgGrantAllowance) (*MsgGrantAllowanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GrantAllowance not implemented")
}
func (UnimplementedMsgServer) RevokeAllowance(context.Context, *MsgRevokeAllowance) (*MsgRevokeAllowanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RevokeAllowance not implemented")
}
func (UnimplementedMsgServer) PruneAllowances(context.Context, *MsgPruneAllowances) (*MsgPruneAllowancesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PruneAllowances not implemented")
}
func (UnimplementedMsgServer) mustEmbedUnimplementedMsgServer() {}
func (UnimplementedMsgServer) testEmbeddedByValue()             {}

// UnsafeMsgServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MsgServer will
// result in compilation errors.
type UnsafeMsgServer interface {
	mustEmbedUnimplementedMsgServer()
}

func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	// If the following call pancis, it indicates UnimplementedMsgServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Msg_ServiceDesc, srv)
}

func _Msg_GrantAllowance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgGrantAllowance)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).GrantAllowance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_GrantAllowance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).GrantAllowance(ctx, req.(*MsgGrantAllowance))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RevokeAllowance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRevokeAllowance)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RevokeAllowance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_RevokeAllowance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RevokeAllowance(ctx, req.(*MsgRevokeAllowance))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_PruneAllowances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgPruneAllowances)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).PruneAllowances(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_PruneAllowances_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).PruneAllowances(ctx, req.(*MsgPruneAllowances))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cosmos.feegrant.v1beta1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GrantAllowance",
			Handler:    _Msg_GrantAllowance_Handler,
		},
		{
			MethodName: "RevokeAllowance",
			Handler:    _Msg_RevokeAllowance_Handler,
		},
		{
			MethodName: "PruneAllowances",
			Handler:    _Msg_PruneAllowances_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cosmos/feegrant/v1beta1/tx.proto",
}
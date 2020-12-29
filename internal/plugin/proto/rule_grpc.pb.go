// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// TfvetRulePluginClient is the client API for TfvetRulePlugin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TfvetRulePluginClient interface {
	GetRuleInfo(ctx context.Context, in *GetRuleInfoRequest, opts ...grpc.CallOption) (*GetRuleInfoResponse, error)
	ExecuteRule(ctx context.Context, in *ExecuteRuleRequest, opts ...grpc.CallOption) (*ExecuteRuleResponse, error)
}

type tfvetRulePluginClient struct {
	cc grpc.ClientConnInterface
}

func NewTfvetRulePluginClient(cc grpc.ClientConnInterface) TfvetRulePluginClient {
	return &tfvetRulePluginClient{cc}
}

func (c *tfvetRulePluginClient) GetRuleInfo(ctx context.Context, in *GetRuleInfoRequest, opts ...grpc.CallOption) (*GetRuleInfoResponse, error) {
	out := new(GetRuleInfoResponse)
	err := c.cc.Invoke(ctx, "/proto.TfvetRulePlugin/GetRuleInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tfvetRulePluginClient) ExecuteRule(ctx context.Context, in *ExecuteRuleRequest, opts ...grpc.CallOption) (*ExecuteRuleResponse, error) {
	out := new(ExecuteRuleResponse)
	err := c.cc.Invoke(ctx, "/proto.TfvetRulePlugin/ExecuteRule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TfvetRulePluginServer is the server API for TfvetRulePlugin service.
// All implementations must embed UnimplementedTfvetRulePluginServer
// for forward compatibility
type TfvetRulePluginServer interface {
	GetRuleInfo(context.Context, *GetRuleInfoRequest) (*GetRuleInfoResponse, error)
	ExecuteRule(context.Context, *ExecuteRuleRequest) (*ExecuteRuleResponse, error)
	mustEmbedUnimplementedTfvetRulePluginServer()
}

// UnimplementedTfvetRulePluginServer must be embedded to have forward compatible implementations.
type UnimplementedTfvetRulePluginServer struct {
}

func (UnimplementedTfvetRulePluginServer) GetRuleInfo(context.Context, *GetRuleInfoRequest) (*GetRuleInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRuleInfo not implemented")
}
func (UnimplementedTfvetRulePluginServer) ExecuteRule(context.Context, *ExecuteRuleRequest) (*ExecuteRuleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteRule not implemented")
}
func (UnimplementedTfvetRulePluginServer) mustEmbedUnimplementedTfvetRulePluginServer() {}

// UnsafeTfvetRulePluginServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TfvetRulePluginServer will
// result in compilation errors.
type UnsafeTfvetRulePluginServer interface {
	mustEmbedUnimplementedTfvetRulePluginServer()
}

func RegisterTfvetRulePluginServer(s grpc.ServiceRegistrar, srv TfvetRulePluginServer) {
	s.RegisterService(&_TfvetRulePlugin_serviceDesc, srv)
}

func _TfvetRulePlugin_GetRuleInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRuleInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TfvetRulePluginServer).GetRuleInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.TfvetRulePlugin/GetRuleInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TfvetRulePluginServer).GetRuleInfo(ctx, req.(*GetRuleInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TfvetRulePlugin_ExecuteRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteRuleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TfvetRulePluginServer).ExecuteRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.TfvetRulePlugin/ExecuteRule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TfvetRulePluginServer).ExecuteRule(ctx, req.(*ExecuteRuleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TfvetRulePlugin_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.TfvetRulePlugin",
	HandlerType: (*TfvetRulePluginServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRuleInfo",
			Handler:    _TfvetRulePlugin_GetRuleInfo_Handler,
		},
		{
			MethodName: "ExecuteRule",
			Handler:    _TfvetRulePlugin_ExecuteRule_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/plugin/proto/rule.proto",
}

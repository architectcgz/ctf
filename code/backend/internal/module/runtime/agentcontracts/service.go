package agentcontracts

import (
	"context"

	"google.golang.org/grpc"
)

const (
	ServiceName                      = "runtime.agent.v1.RuntimeAgent"
	MethodHealth                     = "/runtime.agent.v1.RuntimeAgent/Health"
	MethodCreateNetwork              = "/runtime.agent.v1.RuntimeAgent/CreateNetwork"
	MethodListNetworkSubnets         = "/runtime.agent.v1.RuntimeAgent/ListNetworkSubnets"
	MethodCreateContainer            = "/runtime.agent.v1.RuntimeAgent/CreateContainer"
	MethodResolveServicePort         = "/runtime.agent.v1.RuntimeAgent/ResolveServicePort"
	MethodConnectContainerToNetwork  = "/runtime.agent.v1.RuntimeAgent/ConnectContainerToNetwork"
	MethodInspectContainerNetworkIPs = "/runtime.agent.v1.RuntimeAgent/InspectContainerNetworkIPs"
	MethodStartContainer             = "/runtime.agent.v1.RuntimeAgent/StartContainer"
	MethodStopContainer              = "/runtime.agent.v1.RuntimeAgent/StopContainer"
	MethodRemoveContainer            = "/runtime.agent.v1.RuntimeAgent/RemoveContainer"
	MethodRemoveNetwork              = "/runtime.agent.v1.RuntimeAgent/RemoveNetwork"
	MethodApplyACLRules              = "/runtime.agent.v1.RuntimeAgent/ApplyACLRules"
	MethodApplyACL                   = "/runtime.agent.v1.RuntimeAgent/ApplyACL"
	MethodRemoveACLRules             = "/runtime.agent.v1.RuntimeAgent/RemoveACLRules"
	MethodRemoveACL                  = "/runtime.agent.v1.RuntimeAgent/RemoveACL"
	MethodWriteFileToContainer       = "/runtime.agent.v1.RuntimeAgent/WriteFileToContainer"
	MethodReadFileFromContainer      = "/runtime.agent.v1.RuntimeAgent/ReadFileFromContainer"
	MethodListDirectoryFromContainer = "/runtime.agent.v1.RuntimeAgent/ListDirectoryFromContainer"
	MethodExecContainerCommand       = "/runtime.agent.v1.RuntimeAgent/ExecContainerCommand"
	MethodInspectImageSize           = "/runtime.agent.v1.RuntimeAgent/InspectImageSize"
	MethodRemoveImage                = "/runtime.agent.v1.RuntimeAgent/RemoveImage"
	MethodListManagedContainers      = "/runtime.agent.v1.RuntimeAgent/ListManagedContainers"
	MethodInspectManagedContainer    = "/runtime.agent.v1.RuntimeAgent/InspectManagedContainer"
	MethodListManagedContainerStats  = "/runtime.agent.v1.RuntimeAgent/ListManagedContainerStats"
	MethodRunChecker                 = "/runtime.agent.v1.RuntimeAgent/RunChecker"
	MethodExecContainerInteractive   = "/runtime.agent.v1.RuntimeAgent/ExecContainerInteractive"
)

type RuntimeAgentService interface {
	Health(ctx context.Context, req *HealthRequest) (*HealthResponse, error)
	CreateNetwork(ctx context.Context, req *CreateNetworkRequest) (*CreateNetworkResponse, error)
	ListNetworkSubnets(ctx context.Context, req *ListNetworkSubnetsRequest) (*ListNetworkSubnetsResponse, error)
	CreateContainer(ctx context.Context, req *CreateContainerRequest) (*CreateContainerResponse, error)
	ResolveServicePort(ctx context.Context, req *ResolveServicePortRequest) (*ResolveServicePortResponse, error)
	ConnectContainerToNetwork(ctx context.Context, req *ConnectContainerToNetworkRequest) (*ConnectContainerToNetworkResponse, error)
	InspectContainerNetworkIPs(ctx context.Context, req *InspectContainerNetworkIPsRequest) (*InspectContainerNetworkIPsResponse, error)
	StartContainer(ctx context.Context, req *StartContainerRequest) (*StartContainerResponse, error)
	StopContainer(ctx context.Context, req *StopContainerRequest) (*StopContainerResponse, error)
	RemoveContainer(ctx context.Context, req *RemoveContainerRequest) (*RemoveContainerResponse, error)
	RemoveNetwork(ctx context.Context, req *RemoveNetworkRequest) (*RemoveNetworkResponse, error)
	ApplyACLRules(ctx context.Context, req *ApplyACLRulesRequest) (*ApplyACLRulesResponse, error)
	ApplyACL(ctx context.Context, req *ApplyACLRequest) (*ApplyACLResponse, error)
	RemoveACLRules(ctx context.Context, req *RemoveACLRulesRequest) (*RemoveACLRulesResponse, error)
	RemoveACL(ctx context.Context, req *RemoveACLRequest) (*RemoveACLResponse, error)
	WriteFileToContainer(ctx context.Context, req *WriteFileToContainerRequest) (*WriteFileToContainerResponse, error)
	ReadFileFromContainer(ctx context.Context, req *ReadFileFromContainerRequest) (*ReadFileFromContainerResponse, error)
	ListDirectoryFromContainer(ctx context.Context, req *ListDirectoryFromContainerRequest) (*ListDirectoryFromContainerResponse, error)
	ExecContainerCommand(ctx context.Context, req *ExecContainerCommandRequest) (*ExecContainerCommandResponse, error)
	InspectImageSize(ctx context.Context, req *InspectImageSizeRequest) (*InspectImageSizeResponse, error)
	RemoveImage(ctx context.Context, req *RemoveImageRequest) (*RemoveImageResponse, error)
	ListManagedContainers(ctx context.Context, req *ListManagedContainersRequest) (*ListManagedContainersResponse, error)
	InspectManagedContainer(ctx context.Context, req *InspectManagedContainerRequest) (*InspectManagedContainerResponse, error)
	ListManagedContainerStats(ctx context.Context, req *ListManagedContainerStatsRequest) (*ListManagedContainerStatsResponse, error)
	RunChecker(ctx context.Context, req *RunCheckerRequest) (*RunCheckerResponse, error)
	ExecContainerInteractive(stream RuntimeAgent_ExecContainerInteractiveServer) error
}

type RuntimeAgentClient interface {
	Health(ctx context.Context, req *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error)
	CreateNetwork(ctx context.Context, req *CreateNetworkRequest, opts ...grpc.CallOption) (*CreateNetworkResponse, error)
	ListNetworkSubnets(ctx context.Context, req *ListNetworkSubnetsRequest, opts ...grpc.CallOption) (*ListNetworkSubnetsResponse, error)
	CreateContainer(ctx context.Context, req *CreateContainerRequest, opts ...grpc.CallOption) (*CreateContainerResponse, error)
	ResolveServicePort(ctx context.Context, req *ResolveServicePortRequest, opts ...grpc.CallOption) (*ResolveServicePortResponse, error)
	ConnectContainerToNetwork(ctx context.Context, req *ConnectContainerToNetworkRequest, opts ...grpc.CallOption) (*ConnectContainerToNetworkResponse, error)
	InspectContainerNetworkIPs(ctx context.Context, req *InspectContainerNetworkIPsRequest, opts ...grpc.CallOption) (*InspectContainerNetworkIPsResponse, error)
	StartContainer(ctx context.Context, req *StartContainerRequest, opts ...grpc.CallOption) (*StartContainerResponse, error)
	StopContainer(ctx context.Context, req *StopContainerRequest, opts ...grpc.CallOption) (*StopContainerResponse, error)
	RemoveContainer(ctx context.Context, req *RemoveContainerRequest, opts ...grpc.CallOption) (*RemoveContainerResponse, error)
	RemoveNetwork(ctx context.Context, req *RemoveNetworkRequest, opts ...grpc.CallOption) (*RemoveNetworkResponse, error)
	ApplyACLRules(ctx context.Context, req *ApplyACLRulesRequest, opts ...grpc.CallOption) (*ApplyACLRulesResponse, error)
	ApplyACL(ctx context.Context, req *ApplyACLRequest, opts ...grpc.CallOption) (*ApplyACLResponse, error)
	RemoveACLRules(ctx context.Context, req *RemoveACLRulesRequest, opts ...grpc.CallOption) (*RemoveACLRulesResponse, error)
	RemoveACL(ctx context.Context, req *RemoveACLRequest, opts ...grpc.CallOption) (*RemoveACLResponse, error)
	WriteFileToContainer(ctx context.Context, req *WriteFileToContainerRequest, opts ...grpc.CallOption) (*WriteFileToContainerResponse, error)
	ReadFileFromContainer(ctx context.Context, req *ReadFileFromContainerRequest, opts ...grpc.CallOption) (*ReadFileFromContainerResponse, error)
	ListDirectoryFromContainer(ctx context.Context, req *ListDirectoryFromContainerRequest, opts ...grpc.CallOption) (*ListDirectoryFromContainerResponse, error)
	ExecContainerCommand(ctx context.Context, req *ExecContainerCommandRequest, opts ...grpc.CallOption) (*ExecContainerCommandResponse, error)
	InspectImageSize(ctx context.Context, req *InspectImageSizeRequest, opts ...grpc.CallOption) (*InspectImageSizeResponse, error)
	RemoveImage(ctx context.Context, req *RemoveImageRequest, opts ...grpc.CallOption) (*RemoveImageResponse, error)
	ListManagedContainers(ctx context.Context, req *ListManagedContainersRequest, opts ...grpc.CallOption) (*ListManagedContainersResponse, error)
	InspectManagedContainer(ctx context.Context, req *InspectManagedContainerRequest, opts ...grpc.CallOption) (*InspectManagedContainerResponse, error)
	ListManagedContainerStats(ctx context.Context, req *ListManagedContainerStatsRequest, opts ...grpc.CallOption) (*ListManagedContainerStatsResponse, error)
	RunChecker(ctx context.Context, req *RunCheckerRequest, opts ...grpc.CallOption) (*RunCheckerResponse, error)
	ExecContainerInteractive(ctx context.Context, opts ...grpc.CallOption) (RuntimeAgent_ExecContainerInteractiveClient, error)
}

type runtimeAgentClient struct {
	cc grpc.ClientConnInterface
}

func NewRuntimeAgentClient(cc grpc.ClientConnInterface) RuntimeAgentClient {
	return &runtimeAgentClient{cc: cc}
}

type RuntimeAgent_ExecContainerInteractiveClient interface {
	Send(*ExecContainerInteractiveRequest) error
	Recv() (*ExecContainerInteractiveResponse, error)
	CloseSend() error
	grpc.ClientStream
}

type runtimeAgentExecContainerInteractiveClient struct {
	grpc.ClientStream
}

func (c *runtimeAgentExecContainerInteractiveClient) Send(req *ExecContainerInteractiveRequest) error {
	return c.ClientStream.SendMsg(req)
}

func (c *runtimeAgentExecContainerInteractiveClient) Recv() (*ExecContainerInteractiveResponse, error) {
	resp := new(ExecContainerInteractiveResponse)
	if err := c.ClientStream.RecvMsg(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

type RuntimeAgent_ExecContainerInteractiveServer interface {
	Send(*ExecContainerInteractiveResponse) error
	Recv() (*ExecContainerInteractiveRequest, error)
	grpc.ServerStream
}

type runtimeAgentExecContainerInteractiveServer struct {
	grpc.ServerStream
}

func (s *runtimeAgentExecContainerInteractiveServer) Send(resp *ExecContainerInteractiveResponse) error {
	return s.ServerStream.SendMsg(resp)
}

func (s *runtimeAgentExecContainerInteractiveServer) Recv() (*ExecContainerInteractiveRequest, error) {
	req := new(ExecContainerInteractiveRequest)
	if err := s.ServerStream.RecvMsg(req); err != nil {
		return nil, err
	}
	return req, nil
}

func (c *runtimeAgentClient) Health(ctx context.Context, req *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error) {
	resp := new(HealthResponse)
	if err := c.cc.Invoke(ctx, MethodHealth, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) CreateNetwork(ctx context.Context, req *CreateNetworkRequest, opts ...grpc.CallOption) (*CreateNetworkResponse, error) {
	resp := new(CreateNetworkResponse)
	if err := c.cc.Invoke(ctx, MethodCreateNetwork, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ListNetworkSubnets(ctx context.Context, req *ListNetworkSubnetsRequest, opts ...grpc.CallOption) (*ListNetworkSubnetsResponse, error) {
	resp := new(ListNetworkSubnetsResponse)
	if err := c.cc.Invoke(ctx, MethodListNetworkSubnets, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) CreateContainer(ctx context.Context, req *CreateContainerRequest, opts ...grpc.CallOption) (*CreateContainerResponse, error) {
	resp := new(CreateContainerResponse)
	if err := c.cc.Invoke(ctx, MethodCreateContainer, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ResolveServicePort(ctx context.Context, req *ResolveServicePortRequest, opts ...grpc.CallOption) (*ResolveServicePortResponse, error) {
	resp := new(ResolveServicePortResponse)
	if err := c.cc.Invoke(ctx, MethodResolveServicePort, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ConnectContainerToNetwork(ctx context.Context, req *ConnectContainerToNetworkRequest, opts ...grpc.CallOption) (*ConnectContainerToNetworkResponse, error) {
	resp := new(ConnectContainerToNetworkResponse)
	if err := c.cc.Invoke(ctx, MethodConnectContainerToNetwork, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) InspectContainerNetworkIPs(ctx context.Context, req *InspectContainerNetworkIPsRequest, opts ...grpc.CallOption) (*InspectContainerNetworkIPsResponse, error) {
	resp := new(InspectContainerNetworkIPsResponse)
	if err := c.cc.Invoke(ctx, MethodInspectContainerNetworkIPs, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) StartContainer(ctx context.Context, req *StartContainerRequest, opts ...grpc.CallOption) (*StartContainerResponse, error) {
	resp := new(StartContainerResponse)
	if err := c.cc.Invoke(ctx, MethodStartContainer, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) StopContainer(ctx context.Context, req *StopContainerRequest, opts ...grpc.CallOption) (*StopContainerResponse, error) {
	resp := new(StopContainerResponse)
	if err := c.cc.Invoke(ctx, MethodStopContainer, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) RemoveContainer(ctx context.Context, req *RemoveContainerRequest, opts ...grpc.CallOption) (*RemoveContainerResponse, error) {
	resp := new(RemoveContainerResponse)
	if err := c.cc.Invoke(ctx, MethodRemoveContainer, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) RemoveNetwork(ctx context.Context, req *RemoveNetworkRequest, opts ...grpc.CallOption) (*RemoveNetworkResponse, error) {
	resp := new(RemoveNetworkResponse)
	if err := c.cc.Invoke(ctx, MethodRemoveNetwork, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ApplyACLRules(ctx context.Context, req *ApplyACLRulesRequest, opts ...grpc.CallOption) (*ApplyACLRulesResponse, error) {
	resp := new(ApplyACLRulesResponse)
	if err := c.cc.Invoke(ctx, MethodApplyACLRules, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ApplyACL(ctx context.Context, req *ApplyACLRequest, opts ...grpc.CallOption) (*ApplyACLResponse, error) {
	resp := new(ApplyACLResponse)
	if err := c.cc.Invoke(ctx, MethodApplyACL, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) RemoveACLRules(ctx context.Context, req *RemoveACLRulesRequest, opts ...grpc.CallOption) (*RemoveACLRulesResponse, error) {
	resp := new(RemoveACLRulesResponse)
	if err := c.cc.Invoke(ctx, MethodRemoveACLRules, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) RemoveACL(ctx context.Context, req *RemoveACLRequest, opts ...grpc.CallOption) (*RemoveACLResponse, error) {
	resp := new(RemoveACLResponse)
	if err := c.cc.Invoke(ctx, MethodRemoveACL, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) WriteFileToContainer(ctx context.Context, req *WriteFileToContainerRequest, opts ...grpc.CallOption) (*WriteFileToContainerResponse, error) {
	resp := new(WriteFileToContainerResponse)
	if err := c.cc.Invoke(ctx, MethodWriteFileToContainer, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ReadFileFromContainer(ctx context.Context, req *ReadFileFromContainerRequest, opts ...grpc.CallOption) (*ReadFileFromContainerResponse, error) {
	resp := new(ReadFileFromContainerResponse)
	if err := c.cc.Invoke(ctx, MethodReadFileFromContainer, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ListDirectoryFromContainer(ctx context.Context, req *ListDirectoryFromContainerRequest, opts ...grpc.CallOption) (*ListDirectoryFromContainerResponse, error) {
	resp := new(ListDirectoryFromContainerResponse)
	if err := c.cc.Invoke(ctx, MethodListDirectoryFromContainer, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ExecContainerCommand(ctx context.Context, req *ExecContainerCommandRequest, opts ...grpc.CallOption) (*ExecContainerCommandResponse, error) {
	resp := new(ExecContainerCommandResponse)
	if err := c.cc.Invoke(ctx, MethodExecContainerCommand, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) InspectImageSize(ctx context.Context, req *InspectImageSizeRequest, opts ...grpc.CallOption) (*InspectImageSizeResponse, error) {
	resp := new(InspectImageSizeResponse)
	if err := c.cc.Invoke(ctx, MethodInspectImageSize, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) RemoveImage(ctx context.Context, req *RemoveImageRequest, opts ...grpc.CallOption) (*RemoveImageResponse, error) {
	resp := new(RemoveImageResponse)
	if err := c.cc.Invoke(ctx, MethodRemoveImage, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ListManagedContainers(ctx context.Context, req *ListManagedContainersRequest, opts ...grpc.CallOption) (*ListManagedContainersResponse, error) {
	resp := new(ListManagedContainersResponse)
	if err := c.cc.Invoke(ctx, MethodListManagedContainers, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) InspectManagedContainer(ctx context.Context, req *InspectManagedContainerRequest, opts ...grpc.CallOption) (*InspectManagedContainerResponse, error) {
	resp := new(InspectManagedContainerResponse)
	if err := c.cc.Invoke(ctx, MethodInspectManagedContainer, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ListManagedContainerStats(ctx context.Context, req *ListManagedContainerStatsRequest, opts ...grpc.CallOption) (*ListManagedContainerStatsResponse, error) {
	resp := new(ListManagedContainerStatsResponse)
	if err := c.cc.Invoke(ctx, MethodListManagedContainerStats, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) RunChecker(ctx context.Context, req *RunCheckerRequest, opts ...grpc.CallOption) (*RunCheckerResponse, error) {
	resp := new(RunCheckerResponse)
	if err := c.cc.Invoke(ctx, MethodRunChecker, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *runtimeAgentClient) ExecContainerInteractive(ctx context.Context, opts ...grpc.CallOption) (RuntimeAgent_ExecContainerInteractiveClient, error) {
	desc := &grpc.StreamDesc{ServerStreams: true, ClientStreams: true}
	stream, err := c.cc.NewStream(ctx, desc, MethodExecContainerInteractive, opts...)
	if err != nil {
		return nil, err
	}
	return &runtimeAgentExecContainerInteractiveClient{ClientStream: stream}, nil
}

func RegisterRuntimeAgentService(server grpc.ServiceRegistrar, srv RuntimeAgentService) {
	server.RegisterService(&RuntimeAgentServiceDesc, srv)
}

func unaryMethod[Req any, Resp any](fullMethod string, buildReq func() *Req, invoke func(RuntimeAgentService, context.Context, *Req) (*Resp, error)) grpc.MethodDesc {
	methodName := fullMethod[len("/"+ServiceName+"/"):]
	return grpc.MethodDesc{
		MethodName: methodName,
		Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
			req := buildReq()
			if err := dec(req); err != nil {
				return nil, err
			}
			service := srv.(RuntimeAgentService)
			if interceptor == nil {
				return invoke(service, ctx, req)
			}
			info := &grpc.UnaryServerInfo{
				Server:     srv,
				FullMethod: fullMethod,
			}
			handler := func(ctx context.Context, req any) (any, error) {
				return invoke(service, ctx, req.(*Req))
			}
			return interceptor(ctx, req, info, handler)
		},
	}
}

var RuntimeAgentServiceDesc = grpc.ServiceDesc{
	ServiceName: ServiceName,
	HandlerType: (*RuntimeAgentService)(nil),
	Methods: []grpc.MethodDesc{
		unaryMethod(MethodHealth, func() *HealthRequest { return &HealthRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *HealthRequest) (*HealthResponse, error) {
			return s.Health(ctx, req)
		}),
		unaryMethod(MethodCreateNetwork, func() *CreateNetworkRequest { return &CreateNetworkRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *CreateNetworkRequest) (*CreateNetworkResponse, error) {
			return s.CreateNetwork(ctx, req)
		}),
		unaryMethod(MethodListNetworkSubnets, func() *ListNetworkSubnetsRequest { return &ListNetworkSubnetsRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ListNetworkSubnetsRequest) (*ListNetworkSubnetsResponse, error) {
			return s.ListNetworkSubnets(ctx, req)
		}),
		unaryMethod(MethodCreateContainer, func() *CreateContainerRequest { return &CreateContainerRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *CreateContainerRequest) (*CreateContainerResponse, error) {
			return s.CreateContainer(ctx, req)
		}),
		unaryMethod(MethodResolveServicePort, func() *ResolveServicePortRequest { return &ResolveServicePortRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ResolveServicePortRequest) (*ResolveServicePortResponse, error) {
			return s.ResolveServicePort(ctx, req)
		}),
		unaryMethod(MethodConnectContainerToNetwork, func() *ConnectContainerToNetworkRequest { return &ConnectContainerToNetworkRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ConnectContainerToNetworkRequest) (*ConnectContainerToNetworkResponse, error) {
			return s.ConnectContainerToNetwork(ctx, req)
		}),
		unaryMethod(MethodInspectContainerNetworkIPs, func() *InspectContainerNetworkIPsRequest { return &InspectContainerNetworkIPsRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *InspectContainerNetworkIPsRequest) (*InspectContainerNetworkIPsResponse, error) {
			return s.InspectContainerNetworkIPs(ctx, req)
		}),
		unaryMethod(MethodStartContainer, func() *StartContainerRequest { return &StartContainerRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *StartContainerRequest) (*StartContainerResponse, error) {
			return s.StartContainer(ctx, req)
		}),
		unaryMethod(MethodStopContainer, func() *StopContainerRequest { return &StopContainerRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *StopContainerRequest) (*StopContainerResponse, error) {
			return s.StopContainer(ctx, req)
		}),
		unaryMethod(MethodRemoveContainer, func() *RemoveContainerRequest { return &RemoveContainerRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *RemoveContainerRequest) (*RemoveContainerResponse, error) {
			return s.RemoveContainer(ctx, req)
		}),
		unaryMethod(MethodRemoveNetwork, func() *RemoveNetworkRequest { return &RemoveNetworkRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *RemoveNetworkRequest) (*RemoveNetworkResponse, error) {
			return s.RemoveNetwork(ctx, req)
		}),
		unaryMethod(MethodApplyACLRules, func() *ApplyACLRulesRequest { return &ApplyACLRulesRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ApplyACLRulesRequest) (*ApplyACLRulesResponse, error) {
			return s.ApplyACLRules(ctx, req)
		}),
		unaryMethod(MethodApplyACL, func() *ApplyACLRequest { return &ApplyACLRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ApplyACLRequest) (*ApplyACLResponse, error) {
			return s.ApplyACL(ctx, req)
		}),
		unaryMethod(MethodRemoveACLRules, func() *RemoveACLRulesRequest { return &RemoveACLRulesRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *RemoveACLRulesRequest) (*RemoveACLRulesResponse, error) {
			return s.RemoveACLRules(ctx, req)
		}),
		unaryMethod(MethodRemoveACL, func() *RemoveACLRequest { return &RemoveACLRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *RemoveACLRequest) (*RemoveACLResponse, error) {
			return s.RemoveACL(ctx, req)
		}),
		unaryMethod(MethodWriteFileToContainer, func() *WriteFileToContainerRequest { return &WriteFileToContainerRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *WriteFileToContainerRequest) (*WriteFileToContainerResponse, error) {
			return s.WriteFileToContainer(ctx, req)
		}),
		unaryMethod(MethodReadFileFromContainer, func() *ReadFileFromContainerRequest { return &ReadFileFromContainerRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ReadFileFromContainerRequest) (*ReadFileFromContainerResponse, error) {
			return s.ReadFileFromContainer(ctx, req)
		}),
		unaryMethod(MethodListDirectoryFromContainer, func() *ListDirectoryFromContainerRequest { return &ListDirectoryFromContainerRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ListDirectoryFromContainerRequest) (*ListDirectoryFromContainerResponse, error) {
			return s.ListDirectoryFromContainer(ctx, req)
		}),
		unaryMethod(MethodExecContainerCommand, func() *ExecContainerCommandRequest { return &ExecContainerCommandRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ExecContainerCommandRequest) (*ExecContainerCommandResponse, error) {
			return s.ExecContainerCommand(ctx, req)
		}),
		unaryMethod(MethodInspectImageSize, func() *InspectImageSizeRequest { return &InspectImageSizeRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *InspectImageSizeRequest) (*InspectImageSizeResponse, error) {
			return s.InspectImageSize(ctx, req)
		}),
		unaryMethod(MethodRemoveImage, func() *RemoveImageRequest { return &RemoveImageRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *RemoveImageRequest) (*RemoveImageResponse, error) {
			return s.RemoveImage(ctx, req)
		}),
		unaryMethod(MethodListManagedContainers, func() *ListManagedContainersRequest { return &ListManagedContainersRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ListManagedContainersRequest) (*ListManagedContainersResponse, error) {
			return s.ListManagedContainers(ctx, req)
		}),
		unaryMethod(MethodInspectManagedContainer, func() *InspectManagedContainerRequest { return &InspectManagedContainerRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *InspectManagedContainerRequest) (*InspectManagedContainerResponse, error) {
			return s.InspectManagedContainer(ctx, req)
		}),
		unaryMethod(MethodListManagedContainerStats, func() *ListManagedContainerStatsRequest { return &ListManagedContainerStatsRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *ListManagedContainerStatsRequest) (*ListManagedContainerStatsResponse, error) {
			return s.ListManagedContainerStats(ctx, req)
		}),
		unaryMethod(MethodRunChecker, func() *RunCheckerRequest { return &RunCheckerRequest{} }, func(s RuntimeAgentService, ctx context.Context, req *RunCheckerRequest) (*RunCheckerResponse, error) {
			return s.RunChecker(ctx, req)
		}),
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ExecContainerInteractive",
			Handler:       _RuntimeAgent_ExecContainerInteractive_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
}

func _RuntimeAgent_ExecContainerInteractive_Handler(srv any, stream grpc.ServerStream) error {
	return srv.(RuntimeAgentService).ExecContainerInteractive(&runtimeAgentExecContainerInteractiveServer{ServerStream: stream})
}

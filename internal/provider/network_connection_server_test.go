package provider

import (
	"context"
	"net"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

type getNetworkConnectionMockFunc func(context.Context, *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error)

type fakeNetworkConnectionServiceServer struct {
	network.UnimplementedNetworkConnectionServiceServer

	createMock func(context.Context, *network.CreateNetworkConnectionRequest) (*doublecloud.Operation, error)
	getMock    getNetworkConnectionMockFunc
	deleteMock func(context.Context, *network.DeleteNetworkConnectionRequest) (*doublecloud.Operation, error)
}

func (f *fakeNetworkConnectionServiceServer) Create(ctx context.Context, req *network.CreateNetworkConnectionRequest) (*doublecloud.Operation, error) {
	return f.createMock(ctx, req)
}

func (f *fakeNetworkConnectionServiceServer) Get(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
	return f.getMock(ctx, req)
}

func (f *fakeNetworkConnectionServiceServer) Delete(ctx context.Context, req *network.DeleteNetworkConnectionRequest) (*doublecloud.Operation, error) {
	return f.deleteMock(ctx, req)
}

func startNetworkConnectionServiceMock(f *fakeNetworkConnectionServiceServer) (string, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	gsrv := grpc.NewServer()
	network.RegisterNetworkConnectionServiceServer(gsrv, f)
	fakeServerAddr := l.Addr().String()
	go func() {
		if err := gsrv.Serve(l); err != nil {
			panic(err)
		}
	}()

	return fakeServerAddr, nil
}

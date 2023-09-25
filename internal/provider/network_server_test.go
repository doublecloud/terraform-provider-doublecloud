package provider

import (
	"context"
	"net"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"google.golang.org/grpc"
)

type fakeNetworkServiceServer struct {
	network.UnimplementedNetworkServiceServer

	importMock func(context.Context, *network.ImportNetworkRequest) (*doublecloud.Operation, error)
	getMock    func(context.Context, *network.GetNetworkRequest) (*network.Network, error)
	deleteMock func(context.Context, *network.DeleteNetworkRequest) (*doublecloud.Operation, error)
}

func (f *fakeNetworkServiceServer) Import(ctx context.Context, req *network.ImportNetworkRequest) (*doublecloud.Operation, error) {
	return f.importMock(ctx, req)
}

func (f *fakeNetworkServiceServer) Get(ctx context.Context, req *network.GetNetworkRequest) (*network.Network, error) {
	return f.getMock(ctx, req)
}

func (f *fakeNetworkServiceServer) Delete(ctx context.Context, req *network.DeleteNetworkRequest) (*doublecloud.Operation, error) {
	return f.deleteMock(ctx, req)
}

func startNetworkServiceMock(f *fakeNetworkServiceServer) (string, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	gsrv := grpc.NewServer()
	network.RegisterNetworkServiceServer(gsrv, f)
	fakeServerAddr := l.Addr().String()
	go func() {
		if err := gsrv.Serve(l); err != nil {
			panic(err)
		}
	}()

	return fakeServerAddr, nil
}

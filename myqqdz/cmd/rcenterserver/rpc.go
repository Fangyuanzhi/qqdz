package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"myqqdz/base/glog"
	pb "myqqdz/tools/logicProto"
	"net"
	"sync"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
)

type rCenter struct {
	pb.UnimplementedMetaServer
	savedRooms []*pb.Room
	mu         sync.Mutex
}

// GetRoom 匹配-- logic
func (r *rCenter) GetRoom(ctx context.Context, userList *pb.UserList) (*pb.Room, error) {
	room, fl := selectRoom(userList)
	room.WaitNum = int32(len(room.UserList))
	if !fl {
		glog.Error("[roomSelect],未匹配到房间")
		return room, io.EOF
	}
	glog.Error(room)
	return room, nil
}

// NotifyRoom 更新房间信息-- roomServer
func (r *rCenter) NotifyRoom(ctx context.Context, roomList *pb.RoomList) (*pb.Res, error) {
	glog.Info("[NotifyRoom] 获取到传来的roomList", roomList)
	updaterS(roomList)
	res := &pb.Res{
		Status: 0,
		Msg:    "注册完成",
	}

	return res, nil
}

// selectRoom 通过roomServer的负载情况来选择房间
func newServer() *rCenter {
	s := &rCenter{}
	return s
}

func initGrpc() {
	// 开启gRPC监听服务
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		glog.Error("failed to listen: %v", err)
	}

	// 初始化gRPC服务
	var opts []grpc.ServerOption
	if *tls {
		creeds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			glog.Error("Failed to generate credentials: %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creeds)}
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterMetaServer(grpcServer, newServer())

	err = grpcServer.Serve(lis)
	if err != nil {
		glog.Error("[gRpc] ", err)
		return
	}
	fmt.Println("------hello world-----", fmt.Sprintf("localhost:%d", port))
}

package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"myqqdz/base/glog"
	pb "myqqdz/tools/logicProto"
	"strconv"
	"time"
)

/*
*  处理对rcenter的访问请求
 */

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name used to verify the hostname returned by the TLS handshake")
	client             pb.MetaClient
)

func NotifyRoom(client pb.MetaClient, roomList *pb.RoomList) (*pb.Res, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.NotifyRoom(ctx, roomList)
	if err != nil {
		glog.Error("client.NotifyRoom failed: %v", err)
	}
	return res, err
}

func initGRpc() {
	var opts []grpc.DialOption
	if *tls {
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			glog.Error("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(rCenterAddr, opts...)
	if err != nil {
		glog.Error("fail to dial: %v", err)
	}
	//defer conn.Close()
	client = pb.NewMetaClient(conn)

	register() // 进行注册
}

func register() {
	// 进行注册，rid为0即为注册
	roomList := &pb.RoomList{}
	roomList.RoomList = append(roomList.RoomList, &pb.Room{
		Addr: ip + ":" + strconv.FormatInt(port, 10),
		Id:   0,
	})

	res, er := NotifyRoom(client, roomList)
	if er != nil {
		glog.Error("[Notify]", er)
	}
	glog.Error(res)
}

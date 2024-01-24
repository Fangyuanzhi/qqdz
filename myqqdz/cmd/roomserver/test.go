package main

import (
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"math/rand"
	pb "myqqdz/tools/logicProto"
	"net"
	"time"
)

var conn net.Conn
var err error

func testInit() {
	request := &pb.Operator{
		Operator: "init",
		Token:    tnameToken,
		User: &pb.User{
			Name: name,
		},
		Room: troom,
	}
	context, er := proto.Marshal(request)
	if er != nil {
		fmt.Println("protobuf解析失败")
		return
	}

	_, err = conn.Write(context)
	if err != nil {
		fmt.Println("发送失败")
	}
	buf := [50000]byte{}
	for {
		// 接收服务端信息
		n, er := conn.Read(buf[:])
		if er != nil {
			fmt.Println("recv failed, err:", err)
			return
		}
		fmt.Println("收到数据，大小为:", n)
		operator := &pb.InitResponse{}
		if err = proto.Unmarshal(buf[:n], operator); err != nil {
			fmt.Println("解析错误")
			return
		} else {
			fmt.Println(operator.Player)
			tplayer = operator.Player
			tplayer.X = tplayer.Nx
			tplayer.Y = tplayer.Ny
			tplayer.Nx = tplayer.X + float32(rand.Intn(2)-2)
			tplayer.Ny = tplayer.Y + float32(rand.Intn(2)-2)
		}
		break
	}
}

var mx = float32(rand.Intn(3) - 1)
var my = float32(rand.Intn(3) - 1)
var preTime int64

func testUpdate() {
	if time.Now().Unix()-preTime > 8 {
		preTime = time.Now().Unix()
		mx = float32(rand.Intn(3) - 1)
		my = float32(rand.Intn(3) - 1)
	}
	request := &pb.Operator{
		Operator: "update",
		Token:    tnameToken,
		User: &pb.User{
			Name: name,
		},
		Room: troom,
		Act: &pb.Act{
			MoveS: 5,
			MoveX: mx,
			MoveY: my,
			Type:  2,
		},
	}
	context, er := proto.Marshal(request)
	if er != nil {
		fmt.Println("protobuf解析失败")
		return
	}
	fmt.Println("当前操作：", request.Act)
	_, err = conn.Write(context)
	if err != nil {
		fmt.Println("发送失败")
	}
	bufMax := make([]byte, 0)
	buf := [50000]byte{}
	for {
		// 接收服务端信息
		n, err := conn.Read(buf[:])
		if err != nil {
			fmt.Println("recv failed, err:", err)
			return
		}
		bufMax = append(bufMax, buf[:n]...)
		if n == 50000 {
			continue
		}
		fmt.Println("收到数据，大小为:", len(bufMax))
		operator := &pb.UpdateResponse{}
		if err = proto.Unmarshal(bufMax[:len(bufMax)], operator); err != nil {
			fmt.Println("解析错误")
		} else {
			fmt.Println(operator.Players)
			fmt.Println("吐的孢子", operator.Spore)
		}
		break
	}
}

var op, tp int = 1, 1
var name string

func init() {
	flag.IntVar(&op, "op", 1, "选择操作")
	flag.StringVar(&name, "name", "fht", "选择玩家")
}

var tplayer *pb.Ball
var troom *pb.Room
var tnameToken = "J29eP4n768Q6Z78k"
var troomToken = "uMqFCeJ7Iit0o0P6"
var twidth float32 = 48
var theight float32 = 71
var tx, ty, tnx, tny, tsc float32 = 20.9, 59.9, 37.87, 57.6, 20

func test() {
	filePath := "D:\\codeProgram\\qqdz\\myqqdz\\config\\config.json"
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	cb := func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		fmt.Println(string(value))
	}
	jsonparser.ArrayEach(file, cb, "room", "robotName")
}

func main() {
	test()
	return
	preTime = time.Now().Unix()
	fmt.Println(time.Now().Unix(), time.Now().UnixMilli())
	//for i:=0;i<6;i++{
	//	fmt.Println(i,180*i/3,math.Sin(math.Pi*float64(i)/3),math.Cos(math.Pi*float64(i)/3))
	//}
	//os.Exit(1)
	rand.Seed(time.Now().Unix()) //设置随机初始化种子
	flag.Parse()
	tplayer = &pb.Ball{
		X:     tx,
		Y:     ty,
		Nx:    tnx,
		Ny:    tny,
		Score: tsc,
		Id:    10001,
		Type:  0,
	}
	troom = &pb.Room{
		Id:     0,
		Addr:   "127.0.0.1:7001",
		Width:  twidth,
		Height: theight,
	}
	conn, err = net.Dial("tcp", "127.0.0.1:7001")
	fmt.Println(conn.LocalAddr(), "->", conn.RemoteAddr(), name)
	if err != nil {
		fmt.Println("连接失败")
		return
	}
	testInit()
	fmt.Println("-----------初始化完成---------")
	for {
		fmt.Scan(&op)
		fmt.Println("----------start---------")
		time.Sleep(time.Microsecond * 300)
		testUpdate()
	}
	defer conn.Close()
}

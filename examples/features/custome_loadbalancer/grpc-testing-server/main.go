package main

import (
	"app/grpc-testing-server/api/echo"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type EchoServer struct {
	echo.UnimplementedEchoServer
}

func (e *EchoServer) Echo(ctx context.Context, par *echo.EchoRequest) (*echo.EchoResponse, error) {
	fmt.Println(par.GetMsg())
	result := &echo.EchoResponse{
		Msg: par.GetMsg(),
	}
	return result, nil
}

var (
	// 服务地址，端口，染色
	ip    = "127.0.0.1"
	port  = flag.Int("p", 8080, "port")
	color = flag.String("c", "", "color")

	etcd        = []string{"127.0.0.1:2379"}
	lease int64 = 10
)

// 服务信息
type ServerInfo struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Color   string `json:"color"`
}

func main() {
	log.Println("Go")
	flag.Parse()

	// 创建 etcd client
	cfg := clientv3.Config{
		Endpoints:   etcd,
		DialTimeout: time.Minute,
	}
	var etcdClient *clientv3.Client
	var err error
	if etcdClient, err = clientv3.New(cfg); err != nil {
		log.Fatal("create etcd client failed")
	}
	defer etcdClient.Close()

	//设置租约时间
	grant, err := etcdClient.Grant(context.Background(), lease)
	if err != nil {
		log.Fatal("craete grant failed", err)
	}
	// put key
	key := "/service/echo/" + uuid.New().String()

	serverInfo := ServerInfo{
		Address: ip,
		Port:    *port,
		Color:   *color,
	}

	value, err := json.Marshal(serverInfo)
	if err != nil {
		log.Fatal(err)
	}

	_, err = etcdClient.Put(context.Background(), key, string(value), clientv3.WithLease(grant.ID))
	if err != nil {
		log.Fatal("put k-v failed", err)
	}
	// 定时续约
	leaseRespChan, err := etcdClient.KeepAlive(context.Background(), grant.ID)
	if err != nil {
		log.Fatal("keep alive failed", err)
	}
	go func(c <-chan *clientv3.LeaseKeepAliveResponse) {
		for v := range c {
			fmt.Println("续约成功", v.ResponseHeader)
		}
		log.Println("停止续约")
	}(leaseRespChan)

	// 程序退出时，停止续约
	defer func() {
		etcdClient.Revoke(context.Background(), grant.ID)
		log.Println("停止续约")
	}()

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	echo.RegisterEchoServer(s, &EchoServer{})
	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

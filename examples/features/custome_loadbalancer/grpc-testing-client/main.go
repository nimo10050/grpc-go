package main

import (
	"app/grpc-testing-client/api/echo"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

// 实现 resolver.Builder
type EtcdBuilder struct {
	EtcdEndpoints []string
}

// -----------------------------------------------------------------------------

func (b *EtcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	fmt.Printf("build Scheme : %s Authority : %s Endpoint : %s \n", target.Scheme, target.Authority, target.Endpoint)
	// 创建 Etcd 客户端连接
	cfg := clientv3.Config{
		Endpoints:   b.EtcdEndpoints,
		DialTimeout: time.Minute,
	}
	var etcdClient *clientv3.Client
	var err error
	if etcdClient, err = clientv3.New(cfg); err != nil {
		log.Println("create etcd client failed")
		return nil, err
	}
	// 创建 Resolver
	r := ServiceResolver{
		target:     target,
		cc:         cc,
		EtcdClient: etcdClient,
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return &r, nil
}

func (b *EtcdBuilder) Scheme() string {
	return "grpclb"
}

type ServerInfo struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Color   string `json:"color"`
}

// 实现 resolver.Resolver 接口
type ServiceResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	EtcdClient *clientv3.Client
}

func (r *ServiceResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	// 获取节点列表
	resp, err := r.EtcdClient.Get(context.Background(), fmt.Sprintf("/service/%s", r.target.Endpoint), clientv3.WithPrefix())
	if err != nil {
		log.Println("get service error", err)
	}
	var Address = make([]resolver.Address, 0, resp.Count)
	for _, v := range resp.Kvs {
		fmt.Println(string(v.Key), string(v.Value))
		var serverInfo ServerInfo
		if err := json.Unmarshal(v.Value, &serverInfo); err != nil {
			log.Println(err)
			continue
		}
		// 这里为 Addr 实体添加了 Attributes ， 在负载均衡器器使用颜色来分发流量
		attr := attributes.New("color", serverInfo.Color)
		Address = append(Address, resolver.Address{Addr: fmt.Sprintf("%s:%d", serverInfo.Address, serverInfo.Port), Attributes: attr})
	}
	r.cc.UpdateState(resolver.State{
		Addresses: Address,
	})
}

func (r *ServiceResolver) Close() {
	r.EtcdClient.Close()
}

// -----------------------------------------------------------------------------

// 自定义负载均衡器Builder

type ColorBalancerBuilder struct{}

func (cpb ColorBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	var scs = map[string][]balancer.SubConn{}
	for k, v := range info.ReadySCs {
		color, ok := v.Address.Attributes.Value("color").(string)
		if ok {
			scs[color] = append(scs[color], k)
		} else {
			scs[""] = append(scs[""], k)
		}
	}
	return &ColorBalancer{scs: scs}

}

// 负载均衡器
type ColorBalancer struct {
	scs map[string][]balancer.SubConn
}

func (cb *ColorBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	// 获取请求流量颜色
	md, ok := metadata.FromOutgoingContext(info.Ctx)
	color := ""
	if ok {
		colors := md.Get("color")
		if len(colors) != 0 {
			color = colors[0]
		}
	}
	// 选择节点  这里可以添加容错逻辑，带颜色的节点不存在时，使用没有颜色的节点来处理流量
	sclist := cb.scs[color]
	if len(sclist) == 0 {
		return balancer.PickResult{}, errors.New("pick failed")
	}
	return balancer.PickResult{
		SubConn: sclist[0],
	}, nil
}

// -----------------------------------------------------------------------------

func main() {
	fmt.Println("go")

	// 创建 resolver.Builder
	b := &EtcdBuilder{
		EtcdEndpoints: []string{"127.0.0.1:2379"},
	}
	// 注册 naming resolver
	resolver.Register(b)

	// 注册 负载均衡器
	balancer.Register(base.NewBalancerBuilder("color", ColorBalancerBuilder{}, base.Config{HealthCheck: false}))

	// 修改连接的地址，使用自定义的 name resolver
	conn, err := grpc.Dial(fmt.Sprintf("%s:///%s", b.Scheme(), "echo"),
		grpc.WithInsecure(),
		// 配置 loadBalancing 策略
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"color"}`),
	)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	for i := 0; i < 10; i++ {
		c := echo.NewEchoClient(conn)
		ctx, cancle := context.WithCancel(context.Background())
		defer cancle()
		// 为流量添加颜色
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"color": "green"}))
		r, err := c.Echo(ctx, &echo.EchoRequest{Msg: "hello"})
		if err != nil {
			log.Fatalf("echo failed :%#v", err)
		}
		log.Print(r.GetMsg())
		time.Sleep(time.Second)
	}

	for i := 0; i < 10; i++ {
		c := echo.NewEchoClient(conn)
		ctx, cancle := context.WithCancel(context.Background())
		defer cancle()

		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"color": "red"}))
		r, err := c.Echo(ctx, &echo.EchoRequest{Msg: "hello"})
		if err != nil {
			log.Fatalf("echo failed :%#v", err)
		}
		log.Print(r.GetMsg())
		time.Sleep(time.Second)
	}
}

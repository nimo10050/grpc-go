package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/resolver"
	"log"
	"time"
)

func main() {
	fmt.Println("go")

	// 创建 resolver.Builder
	b := &EtcdBuilder{}
	// 注册 naming resolver
	resolver.Register(b)

	// 注册 负载均衡器
	balancer.Register(base.NewBalancerBuilder("color", ColorBalancerBuilder{}, base.Config{HealthCheck: false}))

	// 修改连接的地址，使用自定义的 name resolver
	add := fmt.Sprintf("%s:///%s", b.Scheme(), "echo")
	conn, err := grpc.Dial(add,
		grpc.WithInsecure(),
		// 配置 loadBalancing 策略
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"color"}`),
	)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	for i := 0; i < 10; i++ {
		c := pb.NewGreeterClient(conn)
		ctx, cancle := context.WithCancel(context.Background())
		defer cancle()

		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "zhangsan"})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s", r.GetMessage())
		time.Sleep(time.Second)
	}
}

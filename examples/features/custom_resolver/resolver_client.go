package main

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

func main() {
	conn, err := grpc.Dial("custom:///localhost:5001", grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("conn: ", conn)

}

// 自定义解析器

// CustomResolver 第一步：实现 Resolver 接口
type CustomResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
}

func (*CustomResolver) ResolveNow(resolver.ResolveNowOptions) {
	fmt.Println("ResolveNow")
}

func (*CustomResolver) Close() {

}

// 第二步：实现 Builder 接口

type CustomResolverBuilder struct {
}

func (*CustomResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	fmt.Println("Build")
	return &CustomResolver{}, nil
}

func (*CustomResolverBuilder) Scheme() string {
	return "custom"
}

// 第三步：注册 ResolverBuilder
func init() {
	resolver.Register(&CustomResolverBuilder{})
}

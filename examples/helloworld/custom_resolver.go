package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"log"
	"time"
)

type EtcdBuilder struct{}

// -----------------------------------------------------------------------------

func (b *EtcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	fmt.Printf("build Scheme : %s Authority : %s Endpoint : %s \n", target.Scheme, target.Authority, target.Endpoint)

	// 创建 Resolver
	r := ServiceResolver{cc: cc, target: target}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return &r, nil
}

func (b *EtcdBuilder) Scheme() string {
	return "grpclb"
}

type ServiceResolver struct {
	cc     resolver.ClientConn
	target resolver.Target
}

func (r *ServiceResolver) Close() {
	//TODO implement me
	panic("implement me")
}

func (r *ServiceResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	client := getEtcdClient()
	key := "addr"

	res, err := client.Get(context.Background(), key)
	if err != nil {
		log.Fatalln("Get kv error", err)
	}

	addresses := make([]resolver.Address, 0, res.Count)

	for _, kv := range res.Kvs {
		fmt.Println("key=", string(kv.Key), ", value=", string(kv.Value))
		addresses = append(addresses, resolver.Address{Addr: string(kv.Value)})
	}

	r.cc.UpdateState(resolver.State{Addresses: addresses})

}

func getEtcdClient() *clientv3.Client {
	addresses := []string{"127.0.0.1:2379"}
	dialTimeout := 5000 * time.Millisecond
	config := clientv3.Config{Endpoints: addresses, DialTimeout: dialTimeout}
	client, err := clientv3.New(config)
	if err != nil {
		log.Fatalln("New client errir")
	}
	return client
}

type ColorBalancerBuilder struct {
}

type ColorBalancer struct {
	scs []balancer.SubConn
}

func (cb *ColorBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	md, ok := metadata.FromOutgoingContext(info.Ctx)
	if !ok {
		log.Fatalln("FromOutgoingContext error.")
	}
	log.Println(md)
	return balancer.PickResult{
		SubConn: cb.scs[0],
	}, nil
}

func (cpb ColorBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	var scs []balancer.SubConn
	for k := range info.ReadySCs {
		scs = append(scs, k)
	}
	fmt.Println("Build a ColorBalancer")
	return &ColorBalancer{}
}

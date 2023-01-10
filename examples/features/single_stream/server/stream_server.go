package main

import (
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/features/single_stream"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedSingleStreamServerServer
}

// LotsOfReplies 返回使用多种语言打招呼
func (s *server) LotsOfReplies(in *pb.StreamRequest, stream pb.SingleStreamServer_LotsOfRepliesServer) error {
	words := []string{
		"你好",
		"hello",
		"こんにちは",
		"안녕하세요",
	}

	for _, word := range words {
		data := &pb.StreamResponse{
			Message: word + in.GetName(),
		}
		// 使用Send方法返回多个数据
		if err := stream.Send(data); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":5001")
	if err != nil {
		log.Printf("net listen error, %s", err)
	}

	s := grpc.NewServer()
	pb.RegisterSingleStreamServerServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Printf("net Serve error, %s", err)
	}
}

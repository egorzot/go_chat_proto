package main

import (
	pb "chat/proto"
	"chat/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":7775")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterChatServer(s, server.NewServer())
	reflection.Register(s)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

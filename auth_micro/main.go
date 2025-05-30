package main

import (
	"fmt"
	"log"
	"net"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
	"google.golang.org/grpc"

	auth "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/server"
)

func main() {
	sessionRepo, err := auth.NewSessionRepo()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with redisClient: %v", err))
		return
	}

	// defer sessionRepo.CloseRepo()
	defer func() {
		if err := sessionRepo.CloseRepo(); err != nil {
			fmt.Printf("error closing session repo: %v", err)
		}
	}()

	// defer auth.ClosePostgresConnection(sessionRepo.DB)
	defer func() {
		if err := auth.ClosePostgresConnection(sessionRepo.DB); err != nil {
			fmt.Printf("error closing postgres connection: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalln("cant listet port", err)
	}

	server := grpc.NewServer()

	sessionService := &auth.SessionServiceServerImpl{
		Repo: sessionRepo,
	}

	sessionpb.RegisterSessionServiceServer(server, sessionService)

	fmt.Println("starting server at :8082")
	server.Serve(lis)
}

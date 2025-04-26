package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	query "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/server"
)

func main() {

	postgresClient, err := query.NewQueryRepo()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with postgresClient: %v", err))
		return
	}
	defer query.ClosePostgresConnection(postgresClient.DB)

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalln("cant listen port", err)
	}

	server := grpc.NewServer()

	queryService := &query.QueryServiceServerImpl{
		Repo: postgresClient,
	}

	querypb.RegisterQueryServiceServer(server, queryService)

	fmt.Println("starting server at :8081")
	server.Serve(lis)
}

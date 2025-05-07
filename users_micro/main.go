package main

import (
	"fmt"
	"net"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	userspb "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/repository"
	impl "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/usecase"
	"google.golang.org/grpc"
)

func main() {
	logger, err := logger.NewLogrusLogger("./logs/access.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	postgresClient, err := users.NewUserRepo()
	if err != nil {
		fmt.Printf("Failed to initialize postgres client: %v\n", err)
		return
	}

	listener, err := net.Listen("tcp", ":8085")
	if err != nil {
		fmt.Println(fmt.Errorf("not able to listen on port 8085: %v", err))
		return
	}

	server := grpc.NewServer()

	usersService := &impl.UserServiceServer{
		UserRepo: postgresClient,
		Logger:   logger,
	}

	userspb.RegisterUsersServiceServer(server, usersService)

	fmt.Println("starting server at :8085")
	server.Serve(listener)
}

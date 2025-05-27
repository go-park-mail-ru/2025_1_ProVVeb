package main

import (
	"fmt"
	"net"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/repository"
	impl "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/usecase"
	"google.golang.org/grpc"
)

func main() {
	logger, err := logger.NewLogrusLogger("./logs/access.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	postgresClient, err := profiles.NewUserRepo()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with postgresClient: %v", err))
		return
	}
	defer postgresClient.CloseRepo()

	staticClient, err := profiles.NewStaticRepo()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with staticClient: %v", err))
		return
	}

	listener, err := net.Listen("tcp", ":8083")
	if err != nil {
		fmt.Println(fmt.Errorf("not able to listen port: %v", err))
		return
	}

	server := grpc.NewServer()

	profilesService := &impl.ProfileServiceServer{
		ProfilesRepo: postgresClient,
		StaticRepo:   staticClient,
		Logger:       logger,
	}

	profilespb.RegisterProfilesServiceServer(server, profilesService)

	fmt.Println("starting server at :8083")
	server.Serve(listener)
}

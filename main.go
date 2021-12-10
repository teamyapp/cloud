package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dep"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Println(err)
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		StartWebAPIServer(cfg)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		StartRPCAPIServer(cfg)
	}()
	wg.Wait()
}

func StartWebAPIServer(cfg config.Config) {
	webAPIServer, err := dep.InitWebAPIServer(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Web server started at port %d\n", cfg.WebAPIPort)
	if err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.WebAPIPort), webAPIServer); err != nil {
		panic(err)
	}
}

func StartRPCAPIServer(cfg config.Config) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCAPIPort))
	if err != nil {
		panic(err)
	}

	fmt.Printf("gRPC server started at port %d\n", cfg.GRPCAPIPort)
	server := dep.InitGRPCAPIServer(cfg)
	panic(server.Serve(lis))
}

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dep"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Println(err)
		panic(err)
	}

	StartAPIServer(cfg)
}

func StartAPIServer(cfg config.Config) {
	webAPIServer, err := dep.InitWebAPIServer(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Web server started at port %d\n", cfg.WebAPIPort)
	if err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.WebAPIPort), webAPIServer); err != nil {
		panic(err)
	}
}

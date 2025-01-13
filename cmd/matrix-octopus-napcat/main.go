package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/duo/matrix-octopus-napcat/internal/common"
	"github.com/duo/matrix-octopus-napcat/internal/octopus"
)

func main() {
	config, err := common.LoadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	log, err := config.Logging.Compile()
	if err != nil {
		panic(err)
	}

	service := octopus.NewService(*log, config)
	go service.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Printf("\n")

	service.Stop()
}

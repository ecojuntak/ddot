package main

import (
	"log"

	"github.com/ecojuntak/ddot/config"
	"github.com/ecojuntak/ddot/proxy"
)

func main() {
	proxyConfig, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("failed to load configuration file: %s", err)
	}

	proxyServer, err := proxy.New(proxyConfig)
	if err != nil {
		log.Fatalf("failed to create server: %s", err)
	}

	proxyServer.Run()
}

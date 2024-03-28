package proxy

import (
	"log/slog"

	"github.com/ecojuntak/ddot/config"
	"github.com/ecojuntak/ddot/handler"
	"github.com/ecojuntak/ddot/resolver"
)

type (
	Handler interface {
		Run()
	}
)

type Proxy struct {
	config    config.Config
	udpServer Handler
	tcpServer Handler
}

func New(config config.Config) (Proxy, error) {
	dnsResolver := resolver.NewResolver(config.TargetServerAddress)
	udpServer, err := handler.NewUDP(config.Host, config.Port, dnsResolver)
	if err != nil {
		slog.Error("failed to create UDP server", "error", err)
		return Proxy{}, err
	}

	tcpServer, err := handler.NewTCP(config.Host, config.Port, config.TcpServerTimeout, dnsResolver)
	if err != nil {
		slog.Error("failed to create TCP server", "error", err)
		return Proxy{}, err
	}

	return Proxy{
		udpServer: udpServer,
		tcpServer: tcpServer,
		config:    config,
	}, nil
}

func (p Proxy) Run() {
	if p.config.UdpServerEnabled {
		go p.udpServer.Run()
	}

	p.tcpServer.Run()
}

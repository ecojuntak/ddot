package handler

import (
	"errors"
	"log/slog"
	"net"

	"github.com/ecojuntak/ddot/resolver"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type udpServer struct {
	*net.UDPConn
	dnsResolver dnsResolver
}

func NewUDP(host string, port int, dnsResolver dnsResolver) (udpServer, error) {
	udpAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}

	udpListener, err := net.ListenUDP("udp", &udpAddr)
	if err != nil {
		return udpServer{}, err
	}

	return udpServer{udpListener, dnsResolver}, nil
}

func (u udpServer) Run() {
	logger := slog.With("protocol", "upd")
	defer func() {
		u.Close()
		logger.Info("UDP listener closed")
	}()

	logger.Info("server started", "address", u.LocalAddr().String())

	for {
		buffer := make([]byte, 1024)
		_, requesterAddr, err := u.ReadFromUDP(buffer)
		if err != nil {
			logger.Error("failed reading message", "err", err)
			continue
		}

		go u.handlerUDPRequest(buffer, requesterAddr)
	}
}

func (u udpServer) handlerUDPRequest(buffer []byte, requesterAddr *net.UDPAddr) {
	logger := slog.With("protocol", "upd")

	requestPacket := gopacket.NewPacket(buffer, layers.LayerTypeDNS, gopacket.Default)
	dnsRequestPacket := requestPacket.Layer(layers.LayerTypeDNS)
	request, ok := dnsRequestPacket.(*layers.DNS)
	if !ok {
		logger.Error("failed parsing request data")
		return
	}

	response, err := u.dnsResolver.QueryDNS(request)
	if err != nil && !errors.Is(err, resolver.DNSTypeNotImplemented) {
		logger.Error("failed querying DNS record", "err", err)
		return
	}
	_, err = u.WriteToUDP(response, requesterAddr)
	if err != nil {
		logger.Error("failed writing DNS response", "err", err)
	}
}

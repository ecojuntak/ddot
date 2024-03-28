package handler

import (
	"encoding/binary"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/ecojuntak/ddot/resolver"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type tcpServer struct {
	*net.TCPListener
	dnsResolver dnsResolver
	timeout     int
}

func NewTCP(host string, port int, timeout int, dnsResolver dnsResolver) (tcpServer, error) {
	tcpAddr := net.TCPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}

	tcpListener, err := net.ListenTCP("tcp", &tcpAddr)
	if err != nil {
		return tcpServer{}, err
	}

	return tcpServer{tcpListener, dnsResolver, timeout}, nil
}

func (t tcpServer) Run() {
	logger := slog.With("protocol", "tcp")
	defer func() {
		t.Close()
		logger.Info("connection closed")
	}()

	logger.Info("server started", "address", t.Addr().String())

	for {
		conn, err := t.Accept()
		if err != nil {
			logger.Error("failed accepting connection", "err", err)
			continue
		}

		duration := time.Duration(t.timeout) * time.Second
		err = conn.SetDeadline(time.Now().Add(duration))
		if err != nil {
			logger.Error("failed to set connection deadline", "err", err)
			continue
		}

		go t.handleTCPRequest(conn)
	}
}

func (t tcpServer) handleTCPRequest(conn net.Conn) {
	logger := slog.With("protocol", "tcp")

	defer func() {
		conn.Close()
		logger.Info("connection closed")
		return
	}()

	logger.Info("receiving request", "remoteAddress", conn.RemoteAddr())

	for {
		buffer := make([]byte, 1024)
		responseSize := make([]byte, 2)

		_, err := conn.Read(buffer)
		if err != nil {
			logger.Error("failed reading buffer", "error", err)
			return
		}

		requestPacket := gopacket.NewPacket(buffer[2:], layers.LayerTypeDNS, gopacket.Default)
		dnsRequestPacket := requestPacket.Layer(layers.LayerTypeDNS)
		request, ok := dnsRequestPacket.(*layers.DNS)
		if !ok {
			logger.Error("failed parsing request data")
			return
		}

		response, err := t.dnsResolver.QueryDNS(request)
		if err != nil && !errors.Is(err, resolver.DNSTypeNotImplemented) {
			logger.Error("failed querying DNS record", "err", err)
			return
		}

		binary.BigEndian.PutUint16(responseSize, uint16(len(response)))
		response = append(responseSize, response...)

		_, err = conn.Write(response)
		if err != nil {
			logger.Error("failed writing response", "error", err)
			return
		} else {
			logger.Info("success sending response", "remoteAddress", conn.RemoteAddr())
			break
		}
	}
}

package handler

import (
	"github.com/google/gopacket/layers"
)

type dnsResolver interface {
	QueryDNS(request *layers.DNS) ([]byte, error)
}

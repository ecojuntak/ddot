package resolver

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Resolver struct {
	*net.Resolver
}

func NewResolver(targetServerAddress string) Resolver {
	return Resolver{
		&net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				tlsConfig := &tls.Config{
					InsecureSkipVerify: false,
				}

				return tls.Dial("tcp", targetServerAddress, tlsConfig)
			},
		},
	}
}

func (r Resolver) QueryDNS(request *layers.DNS) ([]byte, error) {
	hostname := string(request.Questions[0].Name)
	dnsType := request.Questions[0].Type
	dnsClass := request.Questions[0].Class

	logger := slog.With("hostname", hostname, "type", dnsType.String())

	logger.Info("querying domain over TLS")

	ctx := context.Background()
	switch dnsType {
	case layers.DNSTypeA:
		records, err := r.queryDNSTypeA(ctx, hostname, dnsClass)
		if err != nil {
			logger.Error("failed to query hostname over TLS", "err", err)
			return nil, err
		}
		return r.constructDNSSuccessResponse(request, records), nil
	case layers.DNSTypeAAAA:
		records, err := r.queryDNSTypeAAAA(ctx, hostname, dnsClass)
		if err != nil {
			logger.Error("failed to query hostname over TLS", "err", err)
			return nil, err
		}
		return r.constructDNSSuccessResponse(request, records), nil
	default:
		logger.Warn("DNS record type not supported")
		return r.constructDNSErrorResponse(request), DNSTypeNotImplemented
	}
}

func (r Resolver) constructDNSErrorResponse(request *layers.DNS) []byte {
	response := request
	response.QR = true
	response.OpCode = layers.DNSOpCodeQuery
	response.ResponseCode = layers.DNSResponseCodeNotImp

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths: true,
	}
	err := response.SerializeTo(buf, opts)
	if err != nil {
		slog.Error("failed to serialize DNS response", "err", err)
		return nil
	}

	return buf.Bytes()
}

func (r Resolver) constructDNSSuccessResponse(request *layers.DNS, answers []layers.DNSResourceRecord) []byte {
	response := request
	response.QR = true
	response.OpCode = layers.DNSOpCodeQuery
	response.Answers = answers
	response.ResponseCode = layers.DNSResponseCodeNoErr

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths: true,
	}
	err := response.SerializeTo(buf, opts)
	if err != nil {
		slog.Error("failed to serialize DNS response", "err", err)
		return nil
	}

	return buf.Bytes()
}

func (r Resolver) queryDNSTypeA(ctx context.Context, hostname string, dnsClass layers.DNSClass) ([]layers.DNSResourceRecord, error) {
	var dnsRecords []layers.DNSResourceRecord
	ips, err := r.LookupIP(ctx, "ip4", hostname)
	if err != nil {
		slog.Error("failed querying hostname over TLS", "hostname", hostname, "type", "A", "error", err)
		return nil, err
	}

	for _, ip := range ips {
		dnsRecords = append(dnsRecords, layers.DNSResourceRecord{
			Name:  []byte(hostname),
			Type:  layers.DNSTypeA,
			Class: dnsClass,
			IP:    ip,
		})
	}

	return dnsRecords, nil
}

func (r Resolver) queryDNSTypeAAAA(ctx context.Context, hostname string, dnsClass layers.DNSClass) ([]layers.DNSResourceRecord, error) {
	var dnsRecords []layers.DNSResourceRecord
	ips, err := r.LookupIP(ctx, "ip6", hostname)
	if err != nil {
		slog.Error("failed querying hostname over TLS", "hostname", hostname, "type", "AAAA", "error", err)
		return nil, err
	}

	for _, ip := range ips {
		dnsRecords = append(dnsRecords, layers.DNSResourceRecord{
			Name:  []byte(hostname),
			Type:  layers.DNSTypeAAAA,
			Class: dnsClass,
			IP:    ip,
		})
	}
	return dnsRecords, nil
}

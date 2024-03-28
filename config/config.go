package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host                string
	Port                int
	TargetServerAddress string
	UdpServerEnabled    bool
	TcpServerTimeout    int
}

func LoadConfig(configFilePath string) (Config, error) {
	err := godotenv.Load(configFilePath)
	if err != nil {
		return Config{}, err
	}
	portValue, err := strconv.ParseInt(os.Getenv("PORT"), 10, 16)
	if err != nil {
		return Config{}, err
	}

	udpServerEnabledValue, err := strconv.ParseBool(os.Getenv("UDP_SERVER_ENABLED"))
	if err != nil {
		return Config{}, err
	}

	tcpServerTimeoutValue, err := strconv.ParseInt(os.Getenv("TCP_SERVER_TIMEOUT"), 10, 16)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Host:                os.Getenv("HOST"),
		Port:                int(portValue),
		TargetServerAddress: os.Getenv("TARGET_SERVER_ADDRESS"),
		UdpServerEnabled:    udpServerEnabledValue,
		TcpServerTimeout:    int(tcpServerTimeoutValue),
	}, nil
}

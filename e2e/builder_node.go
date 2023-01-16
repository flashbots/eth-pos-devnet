package e2e

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	PortUserRPC   = 8545
	PortEngineRPC = 8551
	PortMetrics   = 6060
)

type BuilderClient struct {
	IP  net.IP
	rpc *rpc.Client
	logger log15.Logger
}

func NewBuilderClient(ip string) (*BuilderClient, error) {
	builderIP := net.ParseIP(ip)

	rpcAddress := fmt.Sprintf("http://%v:%d", builderIP, PortUserRPC)
	client := &http.Client{}
	rpcHttpClient, err := rpc.DialHTTPWithClient(rpcAddress, client)
	if err != nil {
		return nil, err
	}

	return &BuilderClient{
		IP:  builderIP,
		rpc: rpcHttpClient,
		logger: log15.Root(),
	}, nil
}

func (b *BuilderClient) GetMetrics() (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%d/debug/metrics", b.IP, PortMetrics))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

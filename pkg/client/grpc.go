package client

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/go-kit/kit/log/level"

	"gitlab.thovnn.vn/core/sen-kit/senlog"

	"github.com/go-kit/kit/log"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc/connectivity"

	"google.golang.org/grpc"
)

var DefaultLogger = logrus.New()

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

type Client interface {
	GetConnection(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	Close()
}

type grpcClient struct {
	grpcEndpoint string
	conns        map[string]*grpc.ClientConn
	mu           sync.Mutex
	logger       log.Logger
}

func NewGRPCClient(logger log.Logger) Client {
	if logger == nil {
		logg, closeFunc := senlog.New(&senlog.Config{})
		defer closeFunc()
		logger = logg
	}
	return &grpcClient{
		conns:  make(map[string]*grpc.ClientConn),
		logger: logger,
	}
}

func (c *grpcClient) GetConnection(serverAddr string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	if strings.TrimSpace(serverAddr) == "" {
		return nil, errors.New("grpc address is empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	conn, ok := c.conns[serverAddr]
	if !ok || conn.GetState() == connectivity.Shutdown {
		conn, err = c.openNewConnection(serverAddr, opts...)
		if err != nil {
			return nil, err
		}
		c.conns[serverAddr] = conn
	}
	return conn, nil
}

func (c *grpcClient) openNewConnection(serverAddr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if strings.TrimSpace(serverAddr) == "" {
		return nil, errors.New("grpc address is empty")
	}

	if len(opts) == 0 {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(serverAddr, opts...)

	if err != nil {
		level.Error(c.logger).Log("msg", fmt.Sprint("new grpc connection to %s error: %s", serverAddr, err.Error()))
		return nil, err
	}

	level.Info(c.logger).Log("msg", fmt.Sprintf("new grpc connection to %s success ", serverAddr))
	return conn, nil
}

func (c *grpcClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, con := range c.conns {
		_ = con.Close()
	}
	c.conns = nil
}

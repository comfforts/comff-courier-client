package client

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	config "github.com/comfforts/comff-config"
	"github.com/comfforts/logger"

	api "github.com/comfforts/comff-courier/api/v1"
)

const DEFAULT_SERVICE_PORT = "55051"
const DEFAULT_SERVICE_HOST = "127.0.0.1"

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

var (
	defaultDialTimeout      = 5 * time.Second
	defaultKeepAlive        = 30 * time.Second
	defaultKeepAliveTimeout = 10 * time.Second
)

const CourierClientContextKey = ContextKey("courier-client")

type ClientOption struct {
	DialTimeout      time.Duration
	KeepAlive        time.Duration
	KeepAliveTimeout time.Duration
	Caller           string
}

type Client interface {
	RegisterCourier(ctx context.Context, req *api.AddCourierRequest, opts ...grpc.CallOption) (*api.AddCourierResponse, error)
	UpdateCourier(ctx context.Context, req *api.UpdateCourierRequest, opts ...grpc.CallOption) (*api.UpdateCourierResponse, error)
	GetCourier(ctx context.Context, req *api.GetCourierRequest, opts ...grpc.CallOption) (*api.GetCourierResponse, error)
	SearchCouriers(ctx context.Context, req *api.SearchCouriersRequest, opts ...grpc.CallOption) (*api.SearchCouriersResponse, error)
	DeleteCourier(ctx context.Context, req *api.DeleteCourierRequest, opts ...grpc.CallOption) (*api.DeleteResponse, error)
	Close() error
}

func NewDefaultClientOption() *ClientOption {
	return &ClientOption{
		DialTimeout:      defaultDialTimeout,
		KeepAlive:        defaultKeepAlive,
		KeepAliveTimeout: defaultKeepAliveTimeout,
	}
}

type courierClient struct {
	logger logger.AppLogger
	client api.CouriersClient
	conn   *grpc.ClientConn
	opts   *ClientOption
}

func NewClient(logger logger.AppLogger, clientOpts *ClientOption) (*courierClient, error) {
	tlsConfig, err := config.SetupTLSConfig(&config.ConfigOpts{
		Target: config.COURIER_CLIENT,
	})
	if err != nil {
		logger.Error("error setting shops client TLS", zap.Error(err))
		return nil, err
	}
	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
	}

	servicePort := os.Getenv("COURIER_SERVICE_PORT")
	if servicePort == "" {
		servicePort = DEFAULT_SERVICE_PORT
	}
	serviceHost := os.Getenv("COURIER_SERVICE_HOST")
	if serviceHost == "" {
		serviceHost = DEFAULT_SERVICE_HOST
	}

	serviceAddr := fmt.Sprintf("%s:%s", serviceHost, servicePort)
	// with load balancer
	// serviceAddr = fmt.Sprintf("%s:///%s", loadbalance.ShopResolverName, serviceAddr)
	// serviceAddr = fmt.Sprintf("%s:///%s", "shops", serviceAddr)

	conn, err := grpc.Dial(serviceAddr, opts...)
	if err != nil {
		logger.Error("client failed to connect", zap.Error(err))
		return nil, err
	}

	client := api.NewCouriersClient(conn)

	return &courierClient{
		client: client,
		logger: logger,
		conn:   conn,
		opts:   clientOpts,
	}, nil
}

func (cc *courierClient) RegisterCourier(ctx context.Context, req *api.AddCourierRequest, opts ...grpc.CallOption) (*api.AddCourierResponse, error) {
	ctx, cancel := cc.contextWithOptions(ctx, cc.opts)
	defer cancel()

	resp, err := cc.client.RegisterCourier(ctx, req)
	if err != nil {
		cc.logger.Error("error registering courier", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (cc *courierClient) UpdateCourier(ctx context.Context, req *api.UpdateCourierRequest, opts ...grpc.CallOption) (*api.UpdateCourierResponse, error) {
	ctx, cancel := cc.contextWithOptions(ctx, cc.opts)
	defer cancel()

	resp, err := cc.client.UpdateCourier(ctx, req)
	if err != nil {
		cc.logger.Error("error updating courier", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (cc *courierClient) GetCourier(ctx context.Context, req *api.GetCourierRequest, opts ...grpc.CallOption) (*api.GetCourierResponse, error) {
	ctx, cancel := cc.contextWithOptions(ctx, cc.opts)
	defer cancel()

	resp, err := cc.client.GetCourier(ctx, req)
	if err != nil {
		cc.logger.Error("error getting courier", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (cc *courierClient) SearchCouriers(ctx context.Context, req *api.SearchCouriersRequest, opts ...grpc.CallOption) (*api.SearchCouriersResponse, error) {
	ctx, cancel := cc.contextWithOptions(ctx, cc.opts)
	defer cancel()

	resp, err := cc.client.SearchCouriers(ctx, req)
	if err != nil {
		cc.logger.Error("error searching courier", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (cc *courierClient) DeleteCourier(ctx context.Context, req *api.DeleteCourierRequest, opts ...grpc.CallOption) (*api.DeleteResponse, error) {
	ctx, cancel := cc.contextWithOptions(ctx, cc.opts)
	defer cancel()

	resp, err := cc.client.DeleteCourier(ctx, req)
	if err != nil {
		cc.logger.Error("error deleting courier", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (cc *courierClient) Close() error {
	if err := cc.conn.Close(); err != nil {
		cc.logger.Error("error closing courier client connection", zap.Error(err))
		return err
	}
	return nil
}

func (cc *courierClient) contextWithOptions(ctx context.Context, opts *ClientOption) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, cc.opts.DialTimeout)
	if cc.opts.Caller != "" {
		md := metadata.New(map[string]string{"service-client": cc.opts.Caller})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx, cancel
}

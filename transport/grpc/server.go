package grpc

import (
	"context"
	"crypto/tls"
	"github.com/zander-84/gull/internal/endpoint"
	"github.com/zander-84/gull/internal/host"
	"github.com/zander-84/gull/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/url"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
)

// Server is a gRPC server wrapper.
type Server struct {
	*grpc.Server
	endpoint *url.URL
	err      error

	network string
	addr    string
	lis     net.Listener
	tlsConf *tls.Config
	health  *health.Server
}
type ServerOption func(o *Server)

// NewServer creates a gRPC server by options.
func NewServer(addr string, opts ...ServerOption) *Server {

	srv := &Server{
		network: "tcp",
		addr:    addr,
	}
	for _, o := range opts {
		o(srv)
	}

	grpcOpts := []grpc.ServerOption{}
	if srv.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(srv.tlsConf)))
	}
	srv.Server = grpc.NewServer(grpcOpts...)

	grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	reflection.Register(srv.Server)

	return srv
}
func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.addr)
		if err != nil {
			s.err = err
			return err
		}
		s.lis = lis
	}
	if s.endpoint == nil {
		addr, err := host.Extract(s.addr, s.lis)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("grpc", s.tlsConf != nil), addr)
	}
	return s.err
}

// Start  the gRPC server.
func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return s.err
	}
	log.Printf("[GRPC] server listening on: %s \n", s.lis.Addr().String())
	//s.health.Resume()
	return s.Server.Serve(s.lis)
}

// Stop the gRPC server.
func (s *Server) Stop(ctx context.Context) error {
	fin := make(chan struct{}, 1)
	go func() {
		s.Server.GracefulStop()
		log.Printf("[GRPC]  GracefulStop on: %s \n", s.lis.Addr().String())
		fin <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		log.Printf("[GRPC]  Stop Err: %e\n", ctx.Err())
		s.Server.Stop()
		return ctx.Err()
	case <-fin:
		return nil
	}
}

// Endpoint return a real address to registry endpoint.
// examples:
//
//	grpc://127.0.0.1:9000?isSecure=false
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, s.err
	}
	return s.endpoint, nil
}

package grpcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Romasmi/anti-bruteforce/internal/usecases"
	"github.com/Romasmi/anti-bruteforce/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Server struct {
	api.UnimplementedAntiBruteforceServer
	logger     Logger
	grpcServer *grpc.Server
	usecases   map[usecases.Type]usecases.Usecase
}

func NewServer(logger Logger, ucs map[usecases.Type]usecases.Usecase) *Server {
	s := &Server{
		logger:   logger,
		usecases: ucs,
	}

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.loggingInterceptor),
	)
	api.RegisterAntiBruteforceServer(s.grpcServer, s)
	reflection.Register(s.grpcServer)

	return s
}

func (s *Server) Hello(ctx context.Context, req *api.HelloRequest) (*api.HelloResponse, error) {
	uc := s.usecases[usecases.Hello]
	res, err := uc.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.HelloResponse{Message: res.(string)}, nil
}

func (s *Server) Healthcheck(ctx context.Context, req *api.HealthcheckRequest) (*api.HealthcheckResponse, error) {
	uc := s.usecases[usecases.Healthcheck]
	res, err := uc.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.HealthcheckResponse{Status: res.(string)}, nil
}

func (s *Server) CheckAuth(ctx context.Context, req *api.CheckAuthRequest) (*api.CheckAuthResponse, error) {
	input := &usecases.CheckAuthInput{
		Login:    req.Login,
		Password: req.Password,
		IP:       req.Ip,
	}
	res, err := s.usecases[usecases.CheckAuth].Do(ctx, input)
	if err != nil {
		return nil, err
	}
	return &api.CheckAuthResponse{Ok: res.(bool)}, nil
}

func (s *Server) ClearRate(ctx context.Context, req *api.ClearRateRequest) (*api.ClearRateResponse, error) {
	input := &usecases.ClearRateInput{
		Login: req.Login,
		IP:    req.Ip,
	}
	_, err := s.usecases[usecases.ClearRate].Do(ctx, input)
	if err != nil {
		return nil, err
	}
	return &api.ClearRateResponse{}, nil
}

func (s *Server) loggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	msg := fmt.Sprintf("method: %s, duration: %s", info.FullMethod, duration)
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s, error: %v", msg, err))
	} else {
		s.logger.Info(msg)
	}

	return resp, err
}

func (s *Server) Start(host, port string) error {
	addr := net.JoinHostPort(host, port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Info("starting grpc server on " + addr)
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

func (s *Server) Stop() {
	s.logger.Info("stopping grpc server")
	s.grpcServer.GracefulStop()
}

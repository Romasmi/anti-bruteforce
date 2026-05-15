package usecases

import (
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateCheckAuth(login, password, ipStr string) error {
	if login == "" {
		return status.Error(codes.InvalidArgument, "login is required")
	}
	if password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if ipStr == "" {
		return status.Error(codes.InvalidArgument, "ip is required")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return status.Error(codes.InvalidArgument, "ip must be a valid IPv4 address")
	}

	return nil
}

func validateClearRate(login, ipStr string) error {
	if login == "" && ipStr == "" {
		return status.Error(codes.InvalidArgument, "login or ip is required")
	}

	if ipStr != "" {
		ip := net.ParseIP(ipStr)
		if ip == nil || ip.To4() == nil {
			return status.Error(codes.InvalidArgument, "ip must be a valid IPv4 address")
		}
	}

	return nil
}

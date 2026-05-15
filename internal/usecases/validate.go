package usecases

import (
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateIdentifier(t IdentifierType, value string) error {
	if t == IdentifierTypeUnspecified {
		return status.Error(codes.InvalidArgument, "type is required")
	}

	if value == "" {
		return status.Error(codes.InvalidArgument, "value is required")
	}

	if t == IdentifierTypeIP {
		ip := net.ParseIP(value)
		if ip == nil || ip.To4() == nil {
			return status.Error(codes.InvalidArgument, "value must be a valid IPv4 address")
		}
	}

	return nil
}

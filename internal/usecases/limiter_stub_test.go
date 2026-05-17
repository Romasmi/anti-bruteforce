package usecases_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// allowAllLimiter is a test stub that always permits requests and no-ops on Reset.
type allowAllLimiter struct{}

func (allowAllLimiter) Allow(_ map[string]string) (bool, error) { return true, nil }
func (allowAllLimiter) Reset(_, _ string) error                 { return nil }

// denyAllLimiter is a test stub that always denies requests.
type denyAllLimiter struct{}

func (denyAllLimiter) Allow(_ map[string]string) (bool, error) { return false, nil }
func (denyAllLimiter) Reset(_, _ string) error                 { return nil }

// requireCode asserts that err is a gRPC InvalidArgument status error.
func requireCode(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

package usecases_test

import (
	"testing"

	"github.com/Romasmi/anti-bruteforce/internal/usecases"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCheckAuthUsecase(t *testing.T) {
	uc := &usecases.CheckAuthUsecase{}
	ctx := t.Context()

	t.Run("valid request", func(t *testing.T) {
		res, err := uc.Do(ctx, &usecases.CheckAuthInput{
			Login:    "user123",
			Password: "password123",
			IP:       "192.168.1.1",
		})
		require.NoError(t, err)
		require.True(t, res.(bool))
	})

	t.Run("empty login", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{
			Login:    "",
			Password: "password123",
			IP:       "192.168.1.1",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("empty password", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{
			Login:    "user123",
			Password: "",
			IP:       "192.168.1.1",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("empty ip", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{
			Login:    "user123",
			Password: "password123",
			IP:       "",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("invalid ip format", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{
			Login:    "user123",
			Password: "password123",
			IP:       "not-an-ip",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("ipv6 rejected", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{
			Login:    "user123",
			Password: "password123",
			IP:       "2001:db8::1",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("wrong request type", func(t *testing.T) {
		_, err := uc.Do(ctx, "not a CheckAuthInput")
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})
}

package usecases_test

import (
	"context"
	"testing"

	"github.com/Romasmi/anti-bruteforce/internal/usecases"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCheckAuthUsecase(t *testing.T) {
	uc := &usecases.CheckAuthUsecase{}
	ctx := context.Background()

	t.Run("valid login", func(t *testing.T) {
		res, err := uc.Do(ctx, &usecases.CheckAuthInput{Type: usecases.IdentifierTypeLogin, Value: "user123"})
		require.NoError(t, err)
		require.True(t, res.(bool))
	})

	t.Run("valid password", func(t *testing.T) {
		res, err := uc.Do(ctx, &usecases.CheckAuthInput{Type: usecases.IdentifierTypePassword, Value: "s3cr3t"})
		require.NoError(t, err)
		require.True(t, res.(bool))
	})

	t.Run("valid ip", func(t *testing.T) {
		res, err := uc.Do(ctx, &usecases.CheckAuthInput{Type: usecases.IdentifierTypeIP, Value: "192.168.1.1"})
		require.NoError(t, err)
		require.True(t, res.(bool))
	})

	t.Run("unspecified type", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{Type: usecases.IdentifierTypeUnspecified, Value: "user123"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("empty value", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{Type: usecases.IdentifierTypeLogin, Value: ""})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("invalid ip format", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{Type: usecases.IdentifierTypeIP, Value: "not-an-ip"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("ipv6 rejected", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{Type: usecases.IdentifierTypeIP, Value: "2001:db8::1"})
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
